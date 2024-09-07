package main

import (
	"log"
	"net/mail"
	"slices"
	"strings"
)

type Contact struct {
	Id    int
	First string
	Last  string
	Email string
}

type Contacts struct {
	contacts []*Contact
}

func NewContacts() *Contacts {
	return &Contacts{contacts: []*Contact{
		{1, "Matej", "Buocik", "matej.buocik@gmail.com"},
		{2, "Julia", "Rumanova", "julka@gmail.com"},
		{3, "Random", "Kontakt", "random@kontakt.sk"},
	}}
}

func (cs *Contacts) All() []*Contact {
	return cs.contacts
}

func (cs *Contacts) Search(q string) []*Contact {
	found := []*Contact{}
	for _, contact := range cs.contacts {
		if strings.Contains(contact.First, q) || strings.Contains(contact.Last, q) || strings.Contains(contact.Email, q) {
			found = append(found, contact)
		}
	}
	return found
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

	if c.Email == "" {
		errors["Email"] = "Email cannot be empty."
	} else if _, err := mail.ParseAddress(c.Email); err != nil {
		errors["Email"] = "Invalid email address."
	} else if slices.ContainsFunc(cs.contacts, func(co *Contact) bool { return co.Email == c.Email && co.Id != c.Id }) {
		errors["Email"] = "Contact with this email already exists."
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

func (cs *Contacts) Delete(id int) {
	cs.contacts = slices.DeleteFunc(cs.contacts, func(c *Contact) bool { return c.Id == id })
}
