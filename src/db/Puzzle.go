package db

import (
	"../model"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

func (session *Session) PuzzleCreate(p *model.Puzzle) (id string, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	gridstr := strings.Join(p.Grid, "")

	// id is a hash of the grid; this automatically dedupes puzzles
	// with the same grid.  We could also consider hashing the clues
	// if we want to allow puzzles with the same grid but different
	// clue sets
	id = fmt.Sprintf("%x", sha256.Sum256([]byte(gridstr)))

	// Delete any existing child data for this puzzle
    // We don't delete the puzzle itself because we want to
    // keep any solutions that already exist
    childTables := [...]string{"clue", "rebus", "circle", "shade"}
    for _, tbl := range childTables {
        stmt, err := tx.Prepare(`delete from ` + tbl + ` where puzzle_id=?`)
        if err != nil {
            return "", err
        }
        _, err = stmt.Exec(id)
    }

	// now add or update the puzzle
	stmt, err := tx.Prepare(`
        insert into puzzle (id, type, title, author, editor, copyright,
        publisher, date, height, width, grid, notepad)
        values (?,?,?,?,?,?,?,?,?,?,?,?) on duplicate key update
        type=VALUES(type), title=VALUES(title), author=VALUES(author),
        editor=VALUES(editor), copyright=VALUES(copyright),
        publisher=VALUES(publisher), date=VALUES(date),
        height=VALUES(height), width=VALUES(width),
        grid=VALUES(grid), notepad=VALUES(notepad)
    `)
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(
		id,
		p.Type,
		p.Title,
		p.Author,
		p.Editor,
		p.Copyright,
		p.Publisher,
		p.Date,
		p.Height,
		p.Width,
		gridstr,
		p.Notepad)
	if err != nil {
		return "", err
	}

	stmt, err = tx.Prepare(`
        insert into clue (puzzle_id, row, col, number, direction,
                          answer, text) values (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return "", err
	}
	for _, clue := range p.Clues {
		_, err = stmt.Exec(id, clue.Row, clue.Col, clue.Number,
			clue.Direction[0] == 'D', clue.Answer, clue.Text)
		if err != nil {
			return "", err
		}
	}

	stmt, err = tx.Prepare(`
        insert into circle (puzzle_id, row, col) values (?, ?, ?)
    `)
	if err != nil {
		return "", err
	}
	for _, circle := range p.Circles {
		_, err = stmt.Exec(id, circle.Row, circle.Col)
		if err != nil {
			return "", err
		}
	}

	tx.Commit()
	return id, nil
}

func (session *Session) PuzzleGetById(id string) (p model.Puzzle, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return p, err
	}
	defer tx.Rollback()

	if len(id) < 64 {
		id = id + "%"
	}

	var gridstr string
	err = tx.QueryRow(
		`select id, type, title, author, editor, copyright, publisher,
         date, height, width, grid, notepad from puzzle where id like ?`, id).Scan(
		&p.Id,
		&p.Type,
		&p.Title,
		&p.Author,
		&p.Editor,
		&p.Copyright,
		&p.Publisher,
		&p.Date,
		&p.Height,
		&p.Width,
		&gridstr,
		&p.Notepad)

	if err != nil {
		return p, err
	}
	id = p.Id
	if len(gridstr) < p.Width*p.Height {
		return p, errors.New("incomplete grid")
	}
	p.Grid = make([]string, p.Height)
	for i, curs := 0, 0; i < p.Height; i, curs = i+1, curs+p.Width {
		p.Grid[i] = gridstr[curs : curs+p.Width]
	}

	rows, err := tx.Query(
		`select row, col, number, direction, answer, text
         from clue where puzzle_id=? order by direction, number`, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		var clue model.Clue
		var direction int = 0
		err := rows.Scan(
			&clue.Row,
			&clue.Col,
			&clue.Number,
			&direction,
			&clue.Answer,
			&clue.Text)
		if err != nil {
			return p, err
		}
		if direction == 0 {
			clue.Direction = "Across"
		} else {
			clue.Direction = "Down"
		}
		p.Clues = append(p.Clues, clue)
	}
	if err = rows.Err(); err != nil {
		return p, err
	}

	rows, err = tx.Query(
		`select row, col from circle where puzzle_id=?`, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		var circle model.Circle
		err := rows.Scan(
			&circle.Row,
			&circle.Col)
		if err != nil {
			return p, err
		}
		p.Circles = append(p.Circles, circle)
	}
	if err = rows.Err(); err != nil {
		return p, err
	}

	return p, nil
}
