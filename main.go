package main

import (
	"html/template"
	"net/http"
)

var indexTemplate = template.Must(template.ParseFiles("index.html"))

func main() {
	http.HandleFunc("/", serveIndex)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	//show the upload form
	if r.Method != "POST" {
		indexTemplate.Execute(w, r.URL.Path)
	}
}
