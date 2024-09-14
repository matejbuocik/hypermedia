package main

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"strconv"
)

type JSONServer struct {
	prefix   string
	contacts *Contacts
}

func (s JSONServer) renderJSON(w http.ResponseWriter, r *http.Request, v any) {
	js, err := json.Marshal(v)
	if err != nil {
		log.Printf("error: renderJSON: %s (%s)", err, r.URL)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func parseJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return false
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1_000_000))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func (s JSONServer) RegisterHandlers() {
	http.HandleFunc("GET "+s.prefix+"contacts", s.getContacts)
	http.HandleFunc("POST "+s.prefix+"contacts", s.createContact)
	http.HandleFunc("DELETE "+s.prefix+"contacts/{id}", s.deleteContact)
}

func (s JSONServer) getContacts(w http.ResponseWriter, r *http.Request) {
	c := s.contacts.All()
	data := struct {
		Contacts []*Contact `json:"contacts"`
	}{
		Contacts: c,
	}
	s.renderJSON(w, r, data)
}

func (s JSONServer) createContact(w http.ResponseWriter, r *http.Request) {
	c := &Contact{}
	if !parseJSON(w, r, c) {
		return
	}
	c.Id = -1

	errors := s.contacts.Add(c)
	if errors != nil {
		data := struct {
			Errors map[string]string `json:"errors"`
		}{
			Errors: errors,
		}
		w.WriteHeader(http.StatusBadRequest)
		s.renderJSON(w, r, data)
		return
	}

	data := struct {
		Contact *Contact `json:"contact"`
	}{
		Contact: c,
	}

	s.renderJSON(w, r, data)
}

func (s JSONServer) deleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	s.contacts.SetDeleted(id)
	s.contacts.DeletePending()
}
