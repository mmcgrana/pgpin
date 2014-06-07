package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/darkhelmet/env"
	_ "github.com/lib/pq"
	"log"
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

func dataCount(query string, args ...interface{}) (int, error) {
	row := dataConn.QueryRow(query, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DB connection.

var dataConn *sql.DB

func dataStart() {
	log.Print("data.start")
	conn, err := sql.Open("postgres", env.String("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	conn.SetMaxOpenConns(20)
	dataConn = conn
}

// Db operations.

func dataDbValidate(db *Db) error {
	err := validateSlug("name", db.Name)
	if err != nil {
		return err
	}
	err = validatePgUrl("url", db.Url)
	if err != nil {
		return err
	}
	sameNamed, err := dataCount("SELECT count(*) FROM dbs WHERE name=$1 and id!=$2 and removed_at IS NULL", db.Name, db.Id)
	if err != nil {
		return err
	}
	if sameNamed > 0 {
		return &PgpinError{
			Id:         "duplicate-db-name",
			Message:    "name is already used by another db",
			HttpStatus: 400,
		}
	}
	return nil
}

func dataDbList() ([]DbSlim, error) {
	res, err := dataConn.Query("SELECT id, name FROM dbs where removed_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	dbs := []DbSlim{}
	for res.Next() {
		db := DbSlim{}
		err = res.Scan(&db.Id, &db.Name)
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func dataDbAdd(name string, url string) (*Db, error) {
	db := &Db{}
	db.Id = dataRandId()
	db.Name = name
	db.Url = url
	db.AddedAt = time.Now()
	db.UpdatedAt = time.Now()
	err := dataDbValidate(db)
	if err == nil {
		_, err = dataConn.Exec("INSERT INTO dbs (id, name, url, added_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			db.Id, db.Name, db.Url, db.AddedAt, db.UpdatedAt)
	}
	return db, err
}

func dataDbGet(id string) (*Db, error) {
	row := dataConn.QueryRow(`SELECT id, name, url, added_at, updated_at FROM dbs WHERE id=$1 AND removed_at IS NULL LIMIT 1`, id)
	db := Db{}
	err := row.Scan(&db.Id, &db.Name, &db.Url, &db.AddedAt, &db.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, &PgpinError{
			Id:         "not-found",
			Message:    "db not found",
			HttpStatus: 404,
		}
	case err != nil:
		return nil, err
	default:
		return &db, nil
	}
}

func dataDbUpdate(db *Db) error {
	err := dataDbValidate(db)
	if err == nil {
		db.UpdatedAt = time.Now()
		_, err = dataConn.Exec("UPDATE dbs SET name=$1, url=$2, added_at=$3, updated_at=$4, removed_at=$5 WHERE id=$6",
			db.Name, db.Url, db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Id)
	}
	return err
}

func dataDbRemove(id string) (*Db, error) {
	db, err := dataDbGet(id)
	if err != nil {
		return nil, err
	}
	numPins, err := dataCount("SELECT count(*) FROM pins WHERE db_id=$1 AND deleted_at IS NULL", db.Id)
	if err != nil {
		return nil, err
	}
	if numPins != 0 {
		return nil, &PgpinError{
			Id:         "removing-db-with-pins",
			Message:    "cannot remove db with pins",
			HttpStatus: 400,
		}
	}
	removedAt := time.Now()
	db.RemovedAt = &removedAt
	err = dataDbUpdate(db)
	return db, err
}

// Pin operations.

func dataPinValidate(pin *Pin) error {
	err := validateSlug("name", pin.Name)
	if err != nil {
		return err
	}
	err = validateNonempty("query", pin.Query)
	if err != nil {
		return err
	}
	_, err = dataDbGet(pin.DbId)
	if err != nil {
		return err
	}
	sameNamed, err := dataCount("SELECT count(*) FROM pins WHERE name=$1 AND id!=$2 and deleted_at IS NULL", pin.Name, pin.Id)
	if err != nil {
		return err
	} else if sameNamed > 0 {
		return &PgpinError{
			Id:         "duplicate-pin-name",
			Message:    "name is already used by another pin",
			HttpStatus: 400,
		}
	}
	return nil
}

func dataPinList() ([]PinSlim, error) {
	res, err := dataConn.Query("SELECT id, name FROM pins where deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	pins := []PinSlim{}
	for res.Next() {
		pin := PinSlim{}
		err = res.Scan(&pin.Id, &pin.Name)
		if err != nil {
			return nil, err
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func dataPinCreate(dbId string, name string, query string) (*Pin, error) {
	pin := &Pin{}
	pin.Id = dataRandId()
	pin.DbId = dbId
	pin.Name = name
	pin.Query = query
	pin.CreatedAt = time.Now()
	pin.UpdatedAt = time.Now()
	err := dataPinValidate(pin)
	if err != nil {
		return nil, err
	}
	_, err = dataConn.Exec("INSERT INTO pins (id, db_id, name, query, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		pin.Id, pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func dataPinGetInternal(queryFrag string, queryVals ...interface{}) (*Pin, error) {
	row := dataConn.QueryRow(`SELECT id, db_id, name, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error FROM pins WHERE deleted_at IS NULL AND `+queryFrag+` LIMIT 1`, queryVals...)
	pin := Pin{}
	err := row.Scan(&pin.Id, &pin.DbId, &pin.Name, &pin.Query, &pin.CreatedAt, &pin.UpdatedAt, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.ResultsFields, &pin.ResultsRows, &pin.ResultsError)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &pin, nil
	}
}

func dataPinGet(id string) (*Pin, error) {
	pin, err := dataPinGetInternal("id=$1", id)
	if err != nil {
		return nil, err
	}
	if pin == nil {
		return nil, &PgpinError{
			Id:         "not-found",
			Message:    "pin not found",
			HttpStatus: 404,
		}
	}
	return pin, nil
}

func dataPinForQuery() (*Pin, error) {
	return dataPinGetInternal("query_started_at IS NULL AND deleted_at IS NULL")
}

func dataPinUpdate(pin *Pin) error {
	err := dataPinValidate(pin)
	if err != nil {
		return err
	}
	pin.UpdatedAt = time.Now()
	_, err = dataConn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, updated_at=$5, query_started_at=$6, query_finished_at=$7, results_fields=$8, results_rows=$9, results_error=$10, deleted_at=$11 WHERE id=$12",
		pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.UpdatedAt, pin.QueryStartedAt, pin.QueryFinishedAt, pin.ResultsFields, pin.ResultsRows, pin.ResultsError, pin.DeletedAt, pin.Id)
	if err != nil {
		return err
	}
	return nil
}

func dataPinDelete(id string) (*Pin, error) {
	pin, err := dataPinGet(id)
	if err != nil {
		return nil, err
	}
	deletedAt := time.Now()
	pin.DeletedAt = &deletedAt
	err = dataPinUpdate(pin)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func dataPinDbUrl(pin *Pin) (string, error) {
	db, err := dataDbGet(pin.DbId)
	if err != nil {
		return "", err
	}
	return db.Url, nil
}
