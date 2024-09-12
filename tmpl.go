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
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i - 1
		},
	}

	const tmplDir = "tmpl/"

	baseLayoutEntries, err := os.ReadDir(tmplDir + "baseLayout/")
	if err != nil {
		panic("error getting templates - " + err.Error())
	}

	for _, entry := range baseLayoutEntries {
		temp := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles(
			tmplDir+"base.html", tmplDir+"baseLayout/"+entry.Name(), tmplDir+"rows.html", tmplDir+"download.html",
		))
		tmpl[strings.TrimSuffix(entry.Name(), ".html")] = temp
	}

	tmpl["rows"] = template.Must(template.New("rows.html").Funcs(funcMap).ParseFiles(
		tmplDir + "rows.html",
	))
}

func RenderTemplate(w http.ResponseWriter, templ string, data any) {
	err := tmpl[templ].Execute(w, data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", 500)
	}
}
