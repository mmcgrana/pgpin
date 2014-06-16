package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var PgConn *sql.DB

func PgStart() {
	log.Print("pg.start")
	connUrl := fmt.Sprintf(
		"%s?application_name=%s&statement_timeout=%d&connect_timeout=%d",
		ConfigDatabaseUrl,
		"pgpin.api",
		ConfigDatabaseStatementTimeout/time.Millisecond,
		ConfigDatabaseConnectTimeout/time.Millisecond)
	conn, err := sql.Open("postgres", connUrl)
	if err != nil {
		panic(err)
	}
	conn.SetMaxOpenConns(ConfigDatabasePoolSize)
	PgConn = conn
}

func PgCount(query string, args ...interface{}) (int, error) {
	row := PgConn.QueryRow(query, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
