package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ContactServer struct {
	contacts *Contacts
}

func main() {
	// TODO postgres + (sqlc, sqlx)? + http/templates + htmx + alpineJS + flowbite?
	hupSigs := make(chan os.Signal, 1)
	signal.Notify(hupSigs, syscall.SIGHUP)
	go func() {
		for {
			<-hupSigs
			log.Println("Reloading templates...")
			ParseTemplates()
		}
	}()

	done := make(chan bool, 1)
	intSigs := make(chan os.Signal, 1)
	signal.Notify(intSigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-intSigs
		log.Print("Stopping...")
		// TODO Uncomment later: os.RemoveAll("exports")
		done <- true
	}()

	if err := os.MkdirAll("exports", fs.FileMode(0777)); err != nil {
		panic("Mkdir: " + err.Error())
	}
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
	http.HandleFunc("POST /contacts", server.deleteContacts) // This should be DELETE, but we need to send IDs in form body
	http.HandleFunc("GET /contacts/download", server.getContactsDownload)
	http.HandleFunc("GET /contacts/file", server.getContactsFile)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Listening on :8080...")
	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()
	<-done
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
		contacts = s.contacts.SearchPaged(q, page, pageSize)
	} else {
		contacts = s.contacts.AllPaged(page, pageSize)
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

func (s ContactServer) deleteContacts(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for _, val := range r.Form["selectedIDs"] {
		id, err := strconv.Atoi(val)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		s.contacts.SetDeleted(id)
	}

	s.contacts.DeletePending()
	http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}

var upgrader = websocket.Upgrader{}

func (s ContactServer) getContactsDownload(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1011, "error"))

	u, err := uuid.NewRandom()
	if err != nil {
		return
	}
	filename := "exports/" + u.String() + ".csv"
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	if err = csvWriter.Write([]string{"first_name", "last_name", "email"}); err != nil {
		return
	}
	for _, c := range s.contacts.All() {
		if err = csvWriter.Write([]string{c.First, c.Last, c.Email}); err != nil {
			return
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte("10"))
	time.Sleep(time.Second)
	conn.WriteMessage(websocket.TextMessage, []byte("80"))
	time.Sleep(time.Second)
	conn.WriteMessage(websocket.TextMessage, []byte("100"))

	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, u.String()))

	// Remove the file after some time
	time.Sleep(30 * time.Second)
	os.RemoveAll(filename)
}

func (s ContactServer) getContactsFile(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("uuid")
	if err := uuid.Validate(u); err != nil {
		http.Error(w, "Bad uuid", http.StatusBadRequest)
		return
	}

	filename := "exports/" + u + ".csv"
	w.Header().Set("Content-Disposition", "attachment; filename=contacts.csv")
	http.ServeFile(w, r, filename)
	os.Remove(filename)
}
