package formats

import (
	"../puzzle"
	"encoding/xml"
	"errors"
)

type Format interface {
	Format(p puzzle.Puzzle) (data []byte, err error)
	Parse(data []byte) (p puzzle.Puzzle, err error)
}

type xpf struct {
}

type Puzzles struct {
	Puzzles []puzzle.Puzzle `xml:"Puzzle"`
}

func (c *xpf) Format(p puzzle.Puzzle) (data []byte, err error) {
	var doc = Puzzles{}
	doc.Puzzles = make([]puzzle.Puzzle, 1)
	doc.Puzzles[0] = p
	return xml.Marshal(doc)
}

func (c *xpf) Parse(data []byte) (p puzzle.Puzzle, err error) {
	var doc = Puzzles{}
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		return puzzle.Puzzle{}, err
	}
	if len(doc.Puzzles) < 1 {
		return puzzle.Puzzle{}, errors.New("no puzzles found")
	}
	return doc.Puzzles[0], nil
}

func NewXPF() Format {
	return &xpf{}
}
