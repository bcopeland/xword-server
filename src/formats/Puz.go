package formats

import (
	"../model"
	"encoding/binary"
	"encoding/xml"
	"errors"
)

type puz struct {
}

type section struct {
	name string
	data []byte
}

func (c *puz) parseString(ofs int, data []byte) (s string, outofs int) {
	var i int

	// extract iso-8859-1 string, reencode as utf8
	cp := make([]rune, 0, 80)
	for i = ofs; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		cp = append(cp, rune(data[i]))
	}
	return string(cp), i + 1
}

func (c *puz) parseSection(ofs int, data []byte) (s section, outofs int, err error) {
	if len(data) < 8 {
		return s, ofs, errors.New("invalid section: too short")
	}
	name := string(data[ofs : ofs+4])
	ofs += 4
	length := int(binary.LittleEndian.Uint16(data[ofs : ofs+2]))
	ofs += 2
	// TODO
	// checksum := binary.LittleEndian.Uint16(data[ofs:ofs+2])
	ofs += 2

	if length > len(data)-ofs {
		return s, ofs, errors.New("invalid section: too short length")
	}
	return section{name, data[ofs : ofs+length]}, ofs+length, nil
}

func (c *puz) gridAnswer(grid []string, x int, y int, black uint8, direction uint8) (str string) {
	if len(grid) == 0 {
		return str
	}

	height := len(grid)
	width := len(grid[0])

	xinc, yinc := 0, 0
	if direction == model.ACROSS {
		xinc = 1
	} else {
		yinc = 1
	}

	for y < height && x < width && grid[y][x] != black {
		str += string(grid[y][x])
		y += yinc
		x += xinc
	}
	return str
}

func (c *puz) numberGrid(grid []string, black uint8) (clues []model.Clue) {

	if len(grid) == 0 {
		return clues
	}

	height := len(grid)
	width := len(grid[0])
	number := 1

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] == black {
				continue
			}
			start_x := ((x == 0 || grid[y][x-1] == black) &&
				(x+1 < width && grid[y][x+1] != black))
			start_y := ((y == 0 || grid[y-1][x] == black) &&
				(y+1 < height && grid[y+1][x] != black))

			if start_x {
				clue := model.Clue{
					x + 1,
					y + 1,
					number,
					"Across",
					c.gridAnswer(grid, x, y, black, model.ACROSS),
					""}
				clues = append(clues, clue)
			}
			if start_y {
				clue := model.Clue{
					x + 1,
					y + 1,
					number,
					"Down",
					c.gridAnswer(grid, x, y, black, model.DOWN),
					""}
				clues = append(clues, clue)
			}
			if start_x || start_y {
				number++
			}
		}
	}
	return clues
}

func (c *puz) Format(p model.Puzzle) (data []byte, err error) {
	var doc = Puzzles{}
	doc.Puzzles = make([]model.Puzzle, 1)
	doc.Puzzles[0] = p
	return xml.Marshal(doc)
}

func (c *puz) Parse(data []byte) (p model.Puzzle, err error) {
	filelen := len(data)

	if filelen < 0x34 {
		return p, errors.New("data is too small")
	}

	// header
	p.Width = int(data[0x2c])
	p.Height = int(data[0x2d])
	numClues := int(binary.LittleEndian.Uint16(data[0x2e : 0x2e+4]))

	if filelen < 0x34+p.Width*p.Height*2 {
		return p, errors.New("data is too small")
	}

	// grid
	ofs := 0x34
	gridstr := string(data[ofs : ofs+p.Width*p.Height])
	for i := 0; i < p.Height; i++ {
		p.Grid = append(p.Grid, gridstr[i*p.Width:(i+1)*p.Width])
	}
	ofs += p.Width * p.Height

	// partially filled solution, skip
	ofs += p.Width * p.Height

	// strings section
	p.Title, ofs = c.parseString(ofs, data)
	p.Author, ofs = c.parseString(ofs, data)
	p.Copyright, ofs = c.parseString(ofs, data)

	// Clues: these are stored in order of the answer numbering, with
	// across clues preceeding down clues of the same number.  So we
	// can just assign them in the same order that numberGrid creates
	// clue objects
	var clues []string
	for i := 0; i < numClues; i++ {
		var clue string
		clue, ofs = c.parseString(ofs, data)
		clues = append(clues, clue)
	}

	_, ofs = c.parseString(ofs, data)

	// addl sections - of note, GEXT includes circle flags
	for {
		section, out_ofs, err := c.parseSection(ofs, data)
		if err != nil {
			break
		}
		if section.name == "GEXT" {
			for i := range(section.data) {
				b := section.data[i]
				if (b & 0x80) != 0 {
					y := (i / p.Width) + 1;
					x := (i % p.Width) + 1;
					c := model.Circle{ y, x }
					p.Circles = append(p.Circles, c)
				}
			}
		}
		ofs = out_ofs
	}

	p.Clues = c.numberGrid(p.Grid, '.')
	for i := 0; i < len(p.Clues); i++ {
		if i < len(clues) {
			p.Clues[i].Text = clues[i]
		}
	}

	return p, nil
}

func NewPuz() Format {
	return &puz{}
}
