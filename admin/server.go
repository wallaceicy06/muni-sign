package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/wallaceicy06/sign-server/config"
)

var templates = template.Must(template.ParseFiles("templates/index.html"))

var port = flag.Int("port", 8080, "the port to serve this webserver")

func main() {
	flag.Parse()
	http.HandleFunc("/", rootHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("Error starting web server on port %d: %v", port, err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	config, err := config.GetConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
	}
	if err := templates.ExecuteTemplate(w, "index.html", config); err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
	}
}
