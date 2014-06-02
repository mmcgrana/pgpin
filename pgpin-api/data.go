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

var dataConn *sql.DB

func dataStart() {
	log("data.start")
	conf := dataMustParseDatabaseUrl(env.String("DATABASE_URL"))
	conn, err := sql.Open("postgres", conf)
	if err != nil {
		panic(err)
	}
	dataConn = conn
}

// Db operations.

func dataDbList() ([]dbSlim, error) {
	res, err := dataConn.Query("SELECT id, name FROM dbs where removed_at IS NULL")
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

func dataDbAdd(name string, url string) (*db, error) {
	if err := dataValidateSlug("name", name); err != nil {
		return nil, err
	}
	if err := dataValidatePgUrl("url", url); err != nil {
		return nil, err
	}
	db := db{}
	db.Id = dataRandId()
	db.Name = name
	db.Url = url
	db.AddedAt = time.Now()
	_, err := dataConn.Exec("INSERT INTO dbs (id, name, url, added_at) VALUES ($1, $2, $3, $4)",
		db.Id, db.Name, db.Url, db.AddedAt)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func dataDbGet(id string) (*db, error) {
	res, err := dataConn.Query(`SELECT id, name, url, added_at FROM dbs WHERE id=$1 AND removed_at IS NULL LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	ok := res.Next()
	if !ok {
		return nil, &pgpinError{
			Id:         "not-found",
			Message:    "db not found",
			HttpStatus: 404,
		}
	}
	db := db{}
	err = res.Scan(&db.Id, &db.Name, &db.Url, &db.AddedAt)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func dataDbUpdate(db *db) (*db, error) {
	_, err := dataConn.Exec("UPDATE dbs SET name=$1, url=$2, added_at=$3, removed_at=$4 WHERE id=$5",
		db.Name, db.Url, db.AddedAt, db.RemovedAt, db.Id)
	return db, err
}

func dataDbRemove(id string) (*db, error) {
	db, err := dataDbGet(id)
	if err != nil {
		return nil, err
	}
	removedAt := time.Now()
	db.RemovedAt = &removedAt
	return dataDbUpdate(db)
}

// Pin operations.

func dataPinList() ([]pinSlim, error) {
	res, err := dataConn.Query("SELECT id, name FROM pins where deleted_at IS NULL")
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
	if _, err := dataDbGet(dbId); err != nil {
		return nil, err
	}
	pin := pin{}
	pin.Id = dataRandId()
	pin.DbId = dbId
	pin.Name = name
	pin.Query = query
	pin.CreatedAt = time.Now()
	_, err := dataConn.Exec("INSERT INTO pins (id, db_id, name, query, created_at) VALUES ($1, $2, $3, $4, $5)",
		pin.Id, pin.DbId, pin.Name, pin.Query, pin.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func dataPinGetInternal(queryFrag string, queryVals ...interface{}) (*pin, error) {
	res, err := dataConn.Query(`SELECT id, db_id, name, query, created_at, query_started_at, query_finished_at, results_fields_json, results_rows_json, results_error FROM pins WHERE deleted_at IS NULL AND `+queryFrag+` LIMIT 1`, queryVals...)
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

func dataPinForQuery() (*pin, error) {
	return dataPinGetInternal("query_started_at IS NULL AND deleted_at IS NULL")
}

func dataPinUpdate(pin *pin) (*pin, error) {
	_, err := dataConn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, query_started_at=$5, query_finished_at=$6, results_fields_json=$7, results_rows_json=$8, results_error=$9, deleted_at=$10 WHERE id=$11",
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
	db, err := dataDbGet(pin.DbId)
	if err != nil {
		return "", err
	}
	return db.Url, nil
}
