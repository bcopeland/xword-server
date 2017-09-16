package model

type Entry struct {
	SolutionId string
	Ordinal    int
	Version    int
	Value      string
}

type Solution struct {
	Id       string
	PuzzleId string
	Version  int
	Entries  []Entry
}

type SolutionMetadata struct {
	Id        string
	PuzzleId  string
	Title     string
	Author    string
	Progress  int
}

func (s *Solution) GridString() string {
	grid := ""
	for i := 0; i < len(s.Entries); i++ {
		grid += s.Entries[i].Value
	}
	return grid
}
