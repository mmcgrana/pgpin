package main

import (
	"time"
)

type PinSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Pin struct {
	Id              string     `json:"id"`
	Name            string     `json:"name"`
	DbId            string     `json:"db_id"`
	Query           string     `json:"query"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	QueryStartedAt  *time.Time `json:"query_started_at"`
	QueryFinishedAt *time.Time `json:"query_finished_at"`
	ResultsFields   PgJson     `json:"results_fields"`
	ResultsRows     PgJson     `json:"results_rows"`
	ResultsError    *string    `json:"results_error"`
	ReservedAt      *time.Time `json:"-"`
	DeletedAt       *time.Time `json:"-"`
}

type DbSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Db struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	AddedAt   time.Time  `json:"added_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RemovedAt *time.Time `json:"-"`
}
