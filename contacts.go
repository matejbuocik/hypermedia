package main

import (
	"log"
	"net/mail"
	"slices"
	"strings"
)

type Contact struct {
	Id      int
	First   string
	Last    string
	Email   string
	Deleted bool
}

type Contacts struct {
	contacts      []*Contact
	deletePending int
}

func NewContacts() *Contacts {
	return &Contacts{contacts: []*Contact{
		{Id: 1, First: "Matej", Last: "Buocik", Email: "matej.buocik@gmail.com"},
		{Id: 2, First: "Julia", Last: "Rumanova", Email: "julka@gmail.com"},
		{Id: 3, First: "Random1", Last: "Kontakt", Email: "random1@kontakt.sk"},
		{Id: 4, First: "Random2", Last: "Kontakt", Email: "random2@kontakt.sk"},
		{Id: 5, First: "Random3", Last: "Kontakt", Email: "random3@kontakt.sk"},
		{Id: 6, First: "Random4", Last: "Kontakt", Email: "random4@kontakt.sk"},
		{Id: 7, First: "Random5", Last: "Kontakt", Email: "random5@kontakt.sk"},
		{Id: 8, First: "Random6", Last: "Kontakt", Email: "random6@kontakt.sk"},
		{Id: 9, First: "Random7", Last: "Kontakt", Email: "random7@kontakt.sk"},
		{Id: 10, First: "Random8", Last: "Kontakt", Email: "random8@kontakt.sk"},
		{Id: 11, First: "Random9", Last: "Kontakt", Email: "random9@kontakt.sk"},
		{Id: 12, First: "Random10", Last: "Kontakt", Email: "random10@kontakt.sk"},
	}}
}

func contactsPaging(c []*Contact, page int, pageSize int) []*Contact {
	if page <= 0 {
		return nil
	}

	start := min((page-1)*pageSize, len(c))
	end := min(page*pageSize, len(c))

	return c[start:end]
}

func (cs *Contacts) All(page int, pageSize int) []*Contact {
	return contactsPaging(cs.contacts, page, pageSize)
}

func (cs *Contacts) Search(q string, page int, pageSize int) []*Contact {
	found := []*Contact{}
	for _, contact := range cs.contacts {
		if strings.Contains(strings.ToLower(contact.First), strings.ToLower(q)) ||
			strings.Contains(strings.ToLower(contact.Last), strings.ToLower(q)) ||
			strings.Contains(strings.ToLower(contact.Email), strings.ToLower(q)) {
			found = append(found, contact)
		}
	}

	return contactsPaging(found, page, pageSize)
}

func (cs *Contacts) CheckEmailForContact(id int, email string) (string, bool) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "Email cannot be empty.", false
	} else if _, err := mail.ParseAddress(email); err != nil {
		return "Invalid email address.", false
	} else if slices.ContainsFunc(cs.contacts, func(co *Contact) bool { return co.Email == email && co.Id != id }) {
		return "Contact with this email already exists.", false
	}
	return "", true
}

func (cs *Contacts) CheckForErrors(c *Contact) map[string]string {
	errors := make(map[string]string)

	c.First = strings.TrimSpace(c.First)
	c.Last = strings.TrimSpace(c.Last)
	c.Email = strings.TrimSpace(c.Email)

	if c.First == "" {
		errors["First"] = "First name cannot be empty."
	}
	if c.Last == "" {
		errors["Last"] = "Last name cannot be empty."
	}
	if err, ok := cs.CheckEmailForContact(c.Id, c.Email); !ok {
		errors["Email"] = err
	}

	if len(errors) != 0 {
		return errors
	}

	return nil
}

func (cs *Contacts) Add(c *Contact) map[string]string {
	errors := cs.CheckForErrors(c)
	if errors != nil {
		return errors
	}

	if len(cs.contacts) > 0 {
		// Contacts are sorted by Id
		// Ids are reused on deleting contacts!
		c.Id = cs.contacts[len(cs.contacts)-1].Id + 1
	} else {
		c.Id = 1
	}

	cs.contacts = append(cs.contacts, c)
	log.Printf("Add contact: %v\n", c)
	return nil
}

func (cs *Contacts) Edit(c *Contact) map[string]string {
	errors := cs.CheckForErrors(c)
	if errors != nil {
		return errors
	}

	existing, index := cs.Find(c.Id)
	if existing != nil {
		log.Printf("Edit contact: %v -> %v\n", existing, c)
		cs.contacts[index] = c
	}

	return nil
}

func (cs *Contacts) Find(id int) (*Contact, int) {
	index, found := slices.BinarySearchFunc(cs.contacts, id, func(c *Contact, id int) int {
		return c.Id - id
	})

	if found {
		return cs.contacts[index], index
	}

	return nil, -1
}

func (cs *Contacts) SetDeleted(id int) {
	_, i := cs.Find(id)
	cs.contacts[i].Deleted = true
	cs.deletePending++
}

func (cs *Contacts) DeletePending() {
	cs.contacts = slices.DeleteFunc(cs.contacts, func(c *Contact) bool { return c.Deleted })
	cs.deletePending = 0
}
