package service

import (
	"../db"
	"errors"
	"log"
)

type SolutionService struct {
	session *db.Session
}

type SolutionMutation struct {
	Id      string
	Version int
	Grid    string
}

func SolutionServiceNew(session *db.Session) (out *SolutionService) {
	return &SolutionService{session}
}

func (s *SolutionService) Update(u *SolutionMutation) (out *SolutionMutation, err error) {

	solution, err := s.session.SolutionGetById(u.Id)
	if err != nil {
		return out, err
	}

	if len(solution.Entries) != len(u.Grid) {
		return out, errors.New("Updated grid length does not match puzzle")
	}

	nextVer := solution.Version + 1
	changed := false
	for i := 0; i < len(u.Grid); i++ {
		// only accept newer cells
		if solution.Entries[i].Version > u.Version {
			continue
		}
		if solution.Entries[i].Value == string(u.Grid[i]) {
			continue
		}

		solution.Entries[i].Value = string(u.Grid[i])
		solution.Entries[i].Version = nextVer
		changed = true
	}
	if changed {
		solution.Version = nextVer
		solution, err = s.session.SolutionUpdate(solution)
		if err != nil {
			return out, err
		}
	}
	log.Printf("changed : %d version %d\n", changed, nextVer)
	out = &SolutionMutation{
		u.Id,
		solution.Version,
		solution.GridString(),
	}
	return out, nil
}
