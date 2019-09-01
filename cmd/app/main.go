package main

import (
	internal "github.com/aldyaz/csgo-roster/internal/http"
	"github.com/aldyaz/csgo-roster/internal/roster"
)

func main() {
	rosterService := roster.NewService()
	s := internal.NewServer(rosterService)
	s.ServeHTTP()
}
