package formats

import (
	"../model"
	"encoding/xml"
	"errors"
)

type Format interface {
	Format(p model.Puzzle) (data []byte, err error)
	Parse(data []byte) (p model.Puzzle, err error)
}

type xpf struct {
}

type Puzzles struct {
	Puzzles []model.Puzzle `xml:"Puzzle"`
}

func (c *xpf) Format(p model.Puzzle) (data []byte, err error) {
	var doc = Puzzles{}
	doc.Puzzles = make([]model.Puzzle, 1)
	doc.Puzzles[0] = p
	return xml.Marshal(doc)
}

func (c *xpf) Parse(data []byte) (p model.Puzzle, err error) {
	var doc = Puzzles{}
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		return model.Puzzle{}, err
	}
	if len(doc.Puzzles) < 1 {
		return model.Puzzle{}, errors.New("no puzzles found")
	}
	return doc.Puzzles[0], nil
}

func NewXPF() Format {
	return &xpf{}
}
