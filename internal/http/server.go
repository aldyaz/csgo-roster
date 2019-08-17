package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/payfazz/go-skeleton/internal/base"
	"github.com/rs/cors"
)

// Server represents the http server of payfazz-agent
type Server struct {
	devMode   bool
	router    *chi.Mux
	responder *Responder
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				s.responder.Error(w, http.StatusInternalServerError, errors.New(rvr.(string)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// getTokenInfo gets the information inside the access token from the request header
// in this example, the request access token is on the "X-Access-Token" header
func getTokenInfo(r *http.Request) map[string]interface{} {
	token := r.Header.Get("X-Access-Token")

	// just example data:
	return map[string]interface{}{
		"token":  token,
		"userId": 1,
		"phone":  "087777777",
	}
}

func (s *Server) authenticator(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.devMode {
			devUserID, err := strconv.Atoi(r.Header.Get("X-Dev-User-Id"))
			if err == nil && devUserID != 0 {
				ctx := context.WithValue(r.Context(), base.KeyUserID, devUserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		tokenInfo := getTokenInfo(r)
		userID := tokenInfo["userId"].(int)
		ctx := context.WithValue(r.Context(), base.KeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) init() {
	// Basic CORS
	//
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	s.router.Use(
		s.recover,
		cors.Handler,
	)

	s.router.Get("/", s.authenticator(func(w http.ResponseWriter, r *http.Request) {
		currentUser := base.CurrentUser(r.Context())
		if currentUser == nil {
			s.responder.Error(w, http.StatusUnauthorized, errors.New("user is not found"))
			return
		}

		s.responder.JSON(w, http.StatusOK, fmt.Sprintf("Hello, user %d!", *currentUser))
	}))
}

// ServeHTTP serves the http requests
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// NewServer create a new payfazz-cashier http server
func NewServer(
	devMode bool,
	responder *Responder,
) *Server {
	s := &Server{
		devMode:   devMode,
		router:    chi.NewRouter(),
		responder: responder,
	}
	s.init()
	return s
}
