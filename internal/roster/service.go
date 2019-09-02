package roster

import (
	"github.com/aldyaz/csgo-roster/internal/data/entity"
)

type IService interface {
	GetRosters() entity.RosterList
}

type Service struct{}

func (s *Service) GetRosters() entity.RosterList {
	data := []*entity.Roster{
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
	return entity.RosterList{Data: data}
}

func NewService() *Service {
	return &Service{}
}
