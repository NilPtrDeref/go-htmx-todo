package models

import "time"

type Todo struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Completed bool      `db:"completed"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Error struct {
	StatusCode int
	Status     string
	Message    string
}
