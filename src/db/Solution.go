package db

type Solution struct {
	Id       string
	PuzzleId string
	Version  int
	Grid     string
}

func (session *Session) SolutionCreate(s *Solution) (id string, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("insert into solution (id, puzzle_id) values (?, ?)")
	if err != nil {
		return "", err
	}

	// try new random ids, abort if we can't get a unique one
	for i := 0; i < 5; i++ {
		id = session.RandString(16)
		_, err := stmt.Exec(id, s.PuzzleId)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	// now update the solution
	stmt, err = tx.Prepare(`update solution set version=1, grid=? where id=?`)
	if err != nil {
		return "", err
	}

	_, err = stmt.Exec(s.Grid, id)
	if err != nil {
		return "", err
	}

	tx.Commit()
	return id, nil
}

func (session *Session) SolutionGetById(id string) (s Solution, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return s, err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`select id, puzzle_id, version, grid from solution where id=?`, id).Scan(
		&s.Id, &s.PuzzleId, &s.Version, &s.Grid)

	return s, err
}

func (session *Session) SolutionUpdate(s Solution) (out Solution, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return s, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        update solution set version=?, grid=? where id=?
    `)
	if err != nil {
		return s, err
	}

	_, err = stmt.Exec(s.Version, s.Grid, s.Id)
	if err != nil {
		return s, err
	}

	tx.Commit()
	return s, nil
}
