package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type ContactServer struct {
	contacts *Contacts
}

func main() {
	// TODO postgres + (sqlc, sqlx)? + http/templates + htmx + alpineJS + flowbite?
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for {
			<-sigs
			log.Println("Reloading templates...")
			ParseTemplates()
		}
	}()

	ParseTemplates()
	server := ContactServer{contacts: NewContacts()}

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/contacts", http.StatusFound) })
	http.HandleFunc("GET /contacts", server.getContacts)
	http.HandleFunc("GET /contacts/new", server.getNewContact)
	http.HandleFunc("POST /contacts/new", server.postNewContact)
	http.HandleFunc("GET /contacts/{id}/edit", server.getEditContact)
	http.HandleFunc("POST /contacts/{id}/edit", server.postEditContact)
	http.HandleFunc("DELETE /contacts/{id}", server.deleteContact)
	http.HandleFunc("GET /contacts/{id}", server.getContact)
	http.HandleFunc("GET /contacts/{id}/email", server.getContactEmail)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s ContactServer) getContacts(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Trigger") != "load" {
		s.contacts.DeletePending()
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	q := r.URL.Query().Get("q")

	var contacts []*Contact
	pageSize := 10
	if q != "" {
		contacts = s.contacts.Search(q, page, pageSize)
	} else {
		contacts = s.contacts.All(page, pageSize)
	}

	data := struct {
		Title    string
		Query    string
		Contacts []*Contact
		Page     int
	}{
		Title:    "Contacts",
		Query:    q,
		Contacts: contacts,
		Page:     page,
	}

	if t := r.Header.Get("HX-Trigger"); t == "search" || t == "load" {
		RenderTemplate(w, "rows", data)
	} else {
		RenderTemplate(w, "contacts", data)
	}
}

func (s ContactServer) getNewContact(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title   string
		Contact Contact
		Errors  map[string]string
	}{
		Title:   "New contact",
		Contact: Contact{Id: -1},
		Errors:  make(map[string]string),
	}
	RenderTemplate(w, "newContact", data)
}

func (s ContactServer) getEditContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	contact, _ := s.contacts.Find(id)
	if contact == nil {
		http.Redirect(w, r, "/contacts/new", http.StatusFound)
		return
	}

	data := struct {
		Title   string
		Contact *Contact
		Errors  map[string]string
	}{
		Title:   "Edit contact",
		Contact: contact,
		Errors:  make(map[string]string),
	}
	RenderTemplate(w, "newContact", data)
}

func (s ContactServer) postNewContact(w http.ResponseWriter, r *http.Request) {
	new := &Contact{
		Id:    -1,
		First: r.FormValue("First"),
		Last:  r.FormValue("Last"),
		Email: r.FormValue("Email"),
	}

	errors := s.contacts.Add(new)
	if errors == nil {
		http.Redirect(w, r, "/contacts", http.StatusFound)
		return
	}

	data := struct {
		Title   string
		Contact *Contact
		Errors  map[string]string
	}{
		Title:   "New contact",
		Contact: new,
		Errors:  errors,
	}
	RenderTemplate(w, "newContact", data)
}

func (s ContactServer) postEditContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("Id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	new := &Contact{
		Id:    id,
		First: r.FormValue("First"),
		Last:  r.FormValue("Last"),
		Email: r.FormValue("Email"),
	}

	errors := s.contacts.Edit(new)
	if errors == nil {
		http.Redirect(w, r, "/contacts", http.StatusFound)
		return
	}

	data := struct {
		Title   string
		Contact *Contact
		Errors  map[string]string
	}{
		Title:   "Edit contact",
		Contact: new,
		Errors:  errors,
	}
	RenderTemplate(w, "newContact", data)
}

func (s ContactServer) deleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	s.contacts.SetDeleted(id)

	if r.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
	}
}

func (s ContactServer) getContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	contact, _ := s.contacts.Find(id)

	data := struct {
		Title   string
		Contact *Contact
	}{
		Title:   "Contact detail",
		Contact: contact,
	}
	RenderTemplate(w, "contact", data)
}

func (s ContactServer) getContactEmail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("Email")

	emailErr, ok := s.contacts.CheckEmailForContact(id, email)
	if ok {
		fmt.Fprint(w, "")
	} else {
		fmt.Fprint(w, emailErr)
	}
}
