package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var tmpl = map[string]*template.Template{}

func ParseTemplates() {
	const tmplDir = "tmpl/"

	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		panic("error getting templates - " + err.Error())
	}

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i - 1
		},
	}

	for _, entry := range entries {
		temp := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles(tmplDir+"base.html", tmplDir+entry.Name()))
		tmpl[strings.TrimSuffix(entry.Name(), ".html")] = temp
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
