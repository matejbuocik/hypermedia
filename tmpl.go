package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

const tmplDir = "tmpl/"

var tmpl = map[string]*template.Template{}

func ParseTemplates() {
	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		panic("error getting templates - " + err.Error())
	}
	for _, entry := range entries {
		tmpl[strings.TrimSuffix(entry.Name(), ".html")] = template.Must(template.ParseFiles(tmplDir+"base.html", tmplDir+entry.Name()))
	}
}

func RenderTemplate(w http.ResponseWriter, templ string, data any) {
	err := tmpl[templ].Execute(w, data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
	}
}
