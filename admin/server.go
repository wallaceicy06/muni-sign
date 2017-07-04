package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/wallaceicy06/muni-sign/admin/config"
)

var templates = template.Must(template.ParseFiles("templates/index.html"))

var configFilePath = flag.String("config_file", "", "the path to the file that stores the configuration for the sign")
var port = flag.Int("port", 8080, "the port to serve this webserver")

type server struct {
	cfg  config.SignConfig
	port int
}

func main() {
	flag.Parse()

	if *configFilePath == "" {
		fmt.Fprintln(os.Stderr, "A config file path is required.")
		os.Exit(1)
	}

	srv := newServer(*port, config.NewFileSignConfig(*configFilePath)).serve()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	srv.Shutdown(context.Background())
	os.Exit(0)
}

func newServer(port int, cfg config.SignConfig) *server {
	return &server{
		port: port,
		cfg:  cfg,
	}
}

func (s *server) serve() *http.Server {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", s.port),
	}

	http.HandleFunc("/", s.rootHandler)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Error serving: %v", err)
		}
	}()

	return srv
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
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
