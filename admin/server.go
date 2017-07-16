package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/wallaceicy06/muni-sign/admin/config"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

const cacheTimeout = 24 * time.Hour

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/home.html"))

var configFilePath = flag.String("config_file", "", "the path to the file that stores the configuration for the sign")
var nbServerAddr = flag.String("nextbus_server", "", "the address of the nextbus server")

var port = flag.Int("port", 8080, "the port to serve this webserver")

// Alias for time.Now facilitate testing.
var timeNow = time.Now

type server struct {
	cfg         config.SignConfig
	port        int
	nbClient    pb.NextbusClient
	agencyCache *agencyCache
}

type agencyCache struct {
	agencies    []*pb.Agency
	lastRefresh time.Time
}

type rootTemplate struct {
	Cfg      *pb.Configuration
	Agencies []*pb.Agency
}

func main() {
	flag.Parse()

	if *configFilePath == "" {
		fmt.Fprintln(os.Stderr, "A config file path is required.")
		flag.Usage()
		os.Exit(1)
	}

	if *nbServerAddr == "" {
		fmt.Fprintln(os.Stderr, "A nextbus server address is required.")
		flag.Usage()
		os.Exit(1)
	}

	conn, err := grpc.Dial(*nbServerAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error communicatign with nextbus server.")
		os.Exit(1)
	}
	nbClient := pb.NewNextbusClient(conn)

	srv := newServer(*port, nbClient, config.NewFileSignConfig(*configFilePath)).serve()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	srv.Shutdown(context.Background())
	os.Exit(0)
}

func newServer(port int, nbClient pb.NextbusClient, cfg config.SignConfig) *server {
	return &server{
		port:        port,
		cfg:         cfg,
		nbClient:    nbClient,
		agencyCache: &agencyCache{},
	}
}

func (s *server) serve() *http.Server {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", s.port),
	}

	http.HandleFunc("/", s.rootHandler)
	http.HandleFunc("/api/config", s.apiConfigHandler)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

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
		renderRoot(&rootTemplate{c, s.getAgencies()}, w)
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
		stopIdStr := r.Form.Get("stopIds")
		stopIds := strings.Fields(stopIdStr)

		c := &pb.Configuration{
			Agency:  agency,
			StopIds: stopIds,
		}
		if err := s.cfg.Put(c); err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
			return
		}
		renderRoot(&rootTemplate{c, s.getAgencies()}, w)
	default:
		http.Error(w, fmt.Sprintf("Unsupported method: %s.", r.Method), http.StatusMethodNotAllowed)
	}

}

func (s *server) apiConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c, err := s.cfg.Get()
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(c); err != nil {
			log.Printf("Error writing config response: %v", err)
		}
	default:
		http.Error(w, fmt.Sprintf("Unsupported method: %s.", r.Method), http.StatusMethodNotAllowed)
	}
}

func (s *server) getAgencies() []*pb.Agency {
	t := timeNow()

	if d := t.Sub(s.agencyCache.lastRefresh); d > (cacheTimeout) {
		log.Printf("Cache expired: pulling new agency list.")
		res, err := s.nbClient.ListAgencies(context.Background(), &pb.ListAgenciesRequest{})
		if err != nil {
			log.Printf("Error refreshing default agencies: %v", err)
			return s.agencyCache.agencies
		}
		s.agencyCache.agencies = res.GetAgencies()
		s.agencyCache.lastRefresh = timeNow()
	} else {
		log.Printf("Cache fresh: pulling agency list from cache.")
	}

	return s.agencyCache.agencies
}

func renderRoot(t *rootTemplate, w http.ResponseWriter) {
	// Make sure that the configuration is not nil so that the server can return
	// an error before rendering the template.
	if t.Cfg == nil {
		http.Error(w, fmt.Sprintf("Internal error: configuration is nil."), http.StatusInternalServerError)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", t); err != nil {
		log.Printf("Problem rendering HTML template: %v", err)
		return
	}
}
