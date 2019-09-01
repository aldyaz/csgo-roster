package roster

import (
	"github.com/aldyaz/csgo-roster/internal/data/roster"
)

type IService interface {
	GetRosters() roster.RosterList
}

type Service struct{}

func (s *Service) GetRosters() roster.RosterList {
	data := []*roster.Roster{
		{
			ID:   0,
			Name: "Stewie2k",
			Role: "Entry Fragger",
		},
		{
			ID:   1,
			Name: "GuardiaN",
			Role: "AWP",
		},
		{
			ID:   0,
			Name: "device",
			Role: "AWP",
		},
	}
	return roster.RosterList{Data: data}
}

func NewService() *Service {
	return &Service{}
}
