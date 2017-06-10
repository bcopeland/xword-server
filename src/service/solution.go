package service

import (
	"../db"
	"../model"
	"errors"
	"log"
)

type SolutionService struct {
	session *db.Session
}

type SolutionMutation struct {
	Id      string
	Version int
	Entries []model.Entry
}

func SolutionServiceNew(session *db.Session) (out *SolutionService) {
	return &SolutionService{session}
}

func (s *SolutionService) Update(u *SolutionMutation) (out *SolutionMutation, err error) {

	solution, err := s.session.SolutionGetById(u.Id)
	if err != nil {
		return out, err
	}

	if len(solution.Entries) != len(u.Entries) {
		return out, errors.New("Updated grid length does not match puzzle")
	}

	nextVer := solution.Version + 1
	changed := false
	for i := 0; i < len(u.Entries); i++ {
		// only accept newer cells
		if solution.Entries[i].Version > u.Entries[i].Version {
			continue
		}
		if solution.Entries[i].Value == u.Entries[i].Value {
			continue
		}

	    log.Printf("update : %d version %d value %s\n", i, u.Entries[i].Version, u.Entries[i].Value)
		solution.Entries[i].Value = u.Entries[i].Value
		solution.Entries[i].Version = u.Entries[i].Version
		changed = true
	}
	if changed {
		solution.Version = nextVer
		solution, err = s.session.SolutionUpdate(solution)
		if err != nil {
			return out, err
		}
	}
	out = &SolutionMutation{
		u.Id,
		solution.Version,
		solution.Entries}
	return out, nil
}
