package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/bmizerany/pq"
	"github.com/darkhelmet/env"
	"time"
)

// Data helpers.

func dataRandId() string {
	num := 6
	bytes := make([]byte, num)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", bytes)
}

func dataMustParseDatabaseUrl(s string) string {
	conf, err := pq.ParseURL(s)
	if err != nil {
		panic(err)
	}
	return conf
}

// DB connection.

var conn *sql.DB

func DataStart() {
	log("data.start")
	conf := dataMustParseDatabaseUrl(env.String("DATABASE_URL"))
	connNew, err := sql.Open("postgres", conf)
	if err != nil {
		panic(err)
	}
	conn = connNew
}

func DataTest() error {
	log("data.test")
	var r int
	err := conn.QueryRow("SELECT 1").Scan(&r)
	if err != nil {
		return err
	}
	return nil
}

// Db types and functions.

type dbSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type db struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	AddedAt   *time.Time `json:"added_at"`
	RemovedAt *time.Time `json:"removed_at"`
}

func dataDbList() ([]dbSlim, error) {
	res, err := conn.Query("SELECT id, name FROM dbs where removed_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	dbs := []dbSlim{}
	for res.Next() {
		db := dbSlim{}
		err = res.Scan(&db.Id, &db.Name)
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

// Pin types and functions.

type pinSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type pin struct {
	Id                string     `json:"id"`
	Name              string     `json:"name"`
	DbId              string     `json:"db_id"`
	Query             string     `json:"query"`
	CreatedAt         time.Time  `json:"created_at"`
	QueryStartedAt    *time.Time `json:"query_started_at"`
	QueryFinishedAt   *time.Time `json:"query_finished_at"`
	ResultsFieldsJson *string    `json:"results_fields_json"`
	ResultsRowsJson   *string    `json:"results_rows_json"`
	ResultsError      *string    `json:"results_error"`
	DeletedAt         *time.Time `json:"-"`
}

func dataPinList() ([]pinSlim, error) {
	res, err := conn.Query("SELECT id, name FROM pins where deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	pins := []pinSlim{}
	for res.Next() {
		pin := pinSlim{}
		err = res.Scan(&pin.Id, &pin.Name)
		if err != nil {
			return nil, err
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func dataPinCreate(dbId string, name string, query string) (*pin, error) {
	if err := dataValidateSlug("name", name); err != nil {
		return nil, err
	}
	if err := dataValidateNonempty("query", query); err != nil {
		return nil, err
	}
	pin := pin{}
	pin.Id = dataRandId()
	pin.DbId = dbId
	pin.Name = name
	pin.Query = query
	pin.CreatedAt = time.Now()
	_, err := conn.Exec("INSERT INTO pins (id, db_id, name, query, created_at) VALUES ($1, $2, $3, $4, $5)",
		pin.Id, pin.DbId, pin.Name, pin.Query, pin.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func dataPinGetInternal(queryFrag string, queryVals ...interface{}) (*pin, error) {
	res, err := conn.Query(`SELECT id, db_id, name, query, created_at, query_started_at, query_finished_at, results_fields_json, results_rows_json, results_error FROM pins WHERE deleted_at IS NULL AND `+queryFrag+` LIMIT 1`, queryVals...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	ok := res.Next()
	if !ok {
		return nil, nil
	}
	pin := pin{}
	err = res.Scan(&pin.Id, &pin.DbId, &pin.Name, &pin.Query, &pin.CreatedAt, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.ResultsFieldsJson, &pin.ResultsRowsJson, &pin.ResultsError)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func dataPinGet(id string) (*pin, error) {
	pin, err := dataPinGetInternal("id=$1", id)
	if err != nil {
		return nil, err
	}
	if pin == nil {
		return nil, &pgpinError{
			Id:         "not-found",
			Message:    "pin not found",
			HttpStatus: 404,
		}
	}
	return pin, nil
}

func dataPinUpdate(pin *pin) (*pin, error) {
	_, err := conn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, query_started_at=$5, query_finished_at=$6, results_fields_json=$7, results_rows_json=$8, results_error=$9, deleted_at=$10 WHERE id=$11",
		pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.QueryStartedAt, pin.QueryFinishedAt, pin.ResultsFieldsJson, pin.ResultsRowsJson, pin.ResultsError, pin.DeletedAt, pin.Id)
	return pin, err
}

func dataPinDelete(id string) (*pin, error) {
	pin, err := dataPinGet(id)
	if err != nil {
		return nil, err
	}
	deletedAt := time.Now()
	pin.DeletedAt = &deletedAt
	return dataPinUpdate(pin)
}

func dataPinDbUrl(pin *pin) (string, error) {
	return "postgres://postgres:secret@127.0.0.1:5432/pgpin", nil
}
