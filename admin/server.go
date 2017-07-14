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
	pb "github.com/wallaceicy06/muni-sign/proto"
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
		flag.Usage()
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
	switch r.Method {
	case http.MethodGet:
		c, err := s.cfg.Get()
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
			return
		}
		renderRoot(c, w)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
			return
		}
		agency := r.Form.Get("agency")
		if agency == "" {
			http.Error(w, "Agency must be provided.", http.StatusBadRequest)
			return
		}
		stopId := r.Form.Get("stopId")
		if stopId == "" {
			http.Error(w, fmt.Sprintf("Stop ID must be provided."), http.StatusBadRequest)
			return
		}

		c := &pb.Configuration{
			Agency:  agency,
			StopIds: []string{stopId},
		}
		if err := s.cfg.Put(c); err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
			return
		}
		renderRoot(c, w)
	default:

		http.Error(w, fmt.Sprintf("Unsupported method: %s.", r.Method), http.StatusMethodNotAllowed)
	}

}

func renderRoot(cfg *pb.Configuration, w http.ResponseWriter) {
	// Make sure that the configuration is not nil so that the server can return
	// an error before rendering the template.
	if cfg == nil {
		http.Error(w, fmt.Sprintf("Internal error: configuration is nil."), http.StatusInternalServerError)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", cfg); err != nil {
		log.Printf("Problem rendering HTML template: %v", err)
		return
	}

}
