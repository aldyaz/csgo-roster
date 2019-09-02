package entity

type RosterList struct {
	Data []*Roster `json:"data"`
}

type Roster struct {
	ID   int    `json:"rosterId"`
	Name string `json:"name"`
	Role string `json:"role"`
}
