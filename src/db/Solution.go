package db

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

func (session *Session) SolutionCreate(s *Solution) (id string, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("insert into solution (id, puzzle_id, version) values (?, ?, ?)")
	if err != nil {
		return "", err
	}

	// try new random ids, abort if we can't get a unique one
	for i := 0; i < 5; i++ {
		id = session.RandString(16)
		_, err := stmt.Exec(id, s.PuzzleId, s.Version)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	// now create the grid entries
	stmt, err = tx.Prepare(`insert into entry (solution_id, ordinal, version, value) values (?,?,?,?)`)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(s.Entries); i++ {
		_, err = stmt.Exec(id, i, 0, s.Entries[i].Value)
		if err != nil {
			return "", err
		}
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
		`select id, puzzle_id, version from solution where id=?`, id).Scan(
		&s.Id, &s.PuzzleId, &s.Version)
	if err != nil {
		return s, err
	}

	rows, err := tx.Query(
		`select solution_id, ordinal, version, value from entry where solution_id=? order by ordinal`, id)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry Entry
		err := rows.Scan(
			&entry.SolutionId,
			&entry.Ordinal,
			&entry.Version,
			&entry.Value)
		if err != nil {
			return s, err
		}
		s.Entries = append(s.Entries, entry)
	}
	if err = rows.Err(); err != nil {
		return s, err
	}

	return s, err
}

func (session *Session) SolutionUpdate(s Solution) (out Solution, err error) {
	tx, err := session.db.Begin()
	if err != nil {
		return s, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        update solution set version=? where id=?
    `)
	if err != nil {
		return s, err
	}

	_, err = stmt.Exec(s.Version, s.Id)
	if err != nil {
		return s, err
	}
	stmt, err = tx.Prepare(
		`update entry set version=?, value=? where solution_id=? and ordinal=?`)
	if err != nil {
		return s, err
	}
	for i := 0; i < len(s.Entries); i++ {
		_, err = stmt.Exec(s.Entries[i].Version, s.Entries[i].Value, s.Id, i)
		if err != nil {
			return s, err
		}
	}

	tx.Commit()
	return s, nil
}

func (s *Solution) GridString() string {
	grid := ""
	for i := 0; i < len(s.Entries); i++ {
		grid += s.Entries[i].Value
	}
	return grid
}
