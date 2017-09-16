package service

import (
	"../db"
	"../model"
	"errors"
	"log"
	"sort"
	"strings"
)

type ByProgress []model.SolutionMetadata

func (s ByProgress) Len() int {
	return len(s)
}
func (s ByProgress) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByProgress) Less(i, j int) bool {
	return s[i].Progress < s[j].Progress
}

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

func (s *SolutionService) Find() (ids []model.SolutionMetadata, err error) {
	ids, err = s.session.SolutionFind()
	if err != nil {
		return ids, err
	}

	m := make(map[string]([]model.SolutionMetadata))

	for k, item := range ids {
		solution, err := s.session.SolutionGetById(item.Id)
		if err != nil {
			return ids, err
		}
		p, err := s.session.PuzzleGetById(solution.PuzzleId)
		if err != nil {
			return ids, err
		}

		gridstr := strings.Join(p.Grid, "")
		solstr := []rune(solution.GridString())
		total := 0
		correct := 0

		for i, ch := range gridstr {
			if ch == '.' {
				continue
			}

			total++
			if i < len(solstr) && solstr[i] == ch {
				correct++
			}
		}
		if total <= 0 {
			total = 1
		}
		ids[k].Progress = correct * 100 / total
		m[p.Id] = append(m[p.Id], ids[k])
	}

	// dedupe based on puzzle id; include highest solution
	result := make([]model.SolutionMetadata, 0, len(ids))
	for _, sols := range m {
		sort.Sort(sort.Reverse(ByProgress(sols)))
		result = append(result, sols[0])
	}

	sort.Sort(ByProgress(result))

	return result, err
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
