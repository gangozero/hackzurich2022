package main

import (
	"database/sql"
	"encoding/json"
)

type server struct {
	db *sql.DB
}

type ResponseGH struct {
	Paths []Path `json:"paths,omitempty"`
}

type Path struct {
	Points json.RawMessage `json:"points,omitempty"`
}

type ResponseDriver struct {
	User  string  `json:"user,omitempty"`
	Score float64 `json:"score,omitempty"`
}

type driverReview struct {
	user   string
	review float64
}
