package db

import (
	"crypto/rand"
	"database/sql"
)

type Session struct {
	db *sql.DB
}

func NewSession(db *sql.DB) *Session {
	return &Session{db: db}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func (session *Session) RandString(n int) string {
	rbytes := make([]byte, n)
	_, err := rand.Read(rbytes)
	if err != nil {
		panic("rand")
	}

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[int(rbytes[i])%len(letters)]
	}
	return string(b)
}
