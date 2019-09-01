package http

import (
	"fmt"
	"github.com/aldyaz/csgo-roster/internal/http/controller"
	"github.com/aldyaz/csgo-roster/internal/http/response"
	"github.com/aldyaz/csgo-roster/internal/roster"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
)

// Server represents the http server
type Server struct {
	rosterController *controller.RosterController
}

func (s *Server) compileRouter() chi.Router {
	router := chi.NewRouter()
	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	newCors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Access-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	router.Use(newCors.Handler)

	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		response.JSON(res, http.StatusOK, func() {
			fmt.Println("Hello World")
		})
	})

	router.Get("/v1/rosters", s.rosterController.GetRosters())

	return router
}

// ServeHTTP serves the http requests
func (s *Server) ServeHTTP() {
	r := s.compileRouter()
	srv := http.Server{Addr: ":8080", Handler: r}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen %s\n", err)
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}

// NewServer create a new http server
func NewServer(rosterService roster.IService) *Server {
	rosterController := controller.NewRosterController(rosterService)
	return &Server{rosterController: rosterController}
}
