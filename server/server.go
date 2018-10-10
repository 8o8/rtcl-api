package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/rs/cors"
)

type server struct {
	config Config
	port   string
	router *mux.Router
	store  *datastore.Datastore
}

type Config struct {
	Port  string
	Token TokenConfig
}

// tokenConfig configures the tokens issued by the server
type TokenConfig struct {
	Issuer     string
	SigningKey string
	HoursTTL   int
}

// NewServer returns a pointer to an initialised server with a connected datastore
func NewServer(cfg Config, store *datastore.Datastore) *server {
	s := &server{
		config: cfg,
		store:  store,
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

// Start fires up the http server
func (s *server) Start() error {

	// Wrap handler with CORS to handle preflight requests
	ch := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}).Handler(s.router)

	return http.ListenAndServe(":"+s.config.Port, ch)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
