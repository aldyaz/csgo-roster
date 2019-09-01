package controller

import (
	"github.com/aldyaz/csgo-roster/internal/http/response"
	"github.com/aldyaz/csgo-roster/internal/roster"
	"net/http"
)

type RosterController struct {
	rosterService roster.IService
}

func (c *RosterController) GetRosters() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		r := c.rosterService.GetRosters()
		response.JSON(res, http.StatusOK, r)
	}
}

func NewRosterController(rosterService roster.IService) *RosterController {
	return &RosterController{rosterService: rosterService}
}
