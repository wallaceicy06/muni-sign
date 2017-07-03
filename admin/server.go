package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/wallaceicy06/muni-sign/admin/config"
)

var templates = template.Must(template.ParseFiles("templates/index.html"))

var configFilePath = flag.String("config_file", "", "the path to the file that stores the configuration for the sign")
var port = flag.Int("port", 8080, "the port to serve this webserver")

type Server struct {
	cfg *config.SignConfig
}

func main() {
	flag.Parse()

	if *configFilePath == "" {
		fmt.Fprintln(os.Stderr, "A config file path is required.")
		os.Exit(1)
	}

	sc := config.NewSignConfig(*configFilePath)
	server := &Server{
		cfg: sc,
	}
	server.serve()
}

func (s *Server) serve() {
	http.HandleFunc("/", s.rootHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("Error starting web server on port %d: %v", port, err)
	}
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	c, err := s.cfg.Get()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", c); err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
		return
	}
}
