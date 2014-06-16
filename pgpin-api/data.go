package main

import (
	"code.google.com/p/go-uuid/uuid"
	"database/sql"
	"fmt"
	"github.com/fernet/fernet-go"
	"github.com/jrallison/go-workers"
	_ "github.com/lib/pq"
	"log"
	"regexp"
	"time"
)

// Constants.

var DataUuidRegexp = regexp.MustCompilePOSIX("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")

// DB connection.

var DataConn *sql.DB

func DataStart() {
	log.Print("data.start")
	connUrl := fmt.Sprintf("%s?application_name=%s&statement_timeout=%d&connect_timeout=%d",
		ConfigDatabaseUrl, "pgpin.api", ConfigDatabaseStatementTimeout/time.Millisecond, ConfigDatabaseConnectTimeout/time.Millisecond)
	conn, err := sql.Open("postgres", connUrl)
	if err != nil {
		panic(err)
	}
	conn.SetMaxOpenConns(ConfigDatabasePoolSize)
	DataConn = conn
}

// Data helpers.

func DataCount(query string, args ...interface{}) (int, error) {
	row := DataConn.QueryRow(query, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func DataFernetEncrypt(s string) []byte {
	tok, err := fernet.EncryptAndSign([]byte(s), ConfigFernetKeys[0])
	Must(err)
	return tok
}

func DataFernetDecrypt(b []byte) string {
	msg := fernet.VerifyAndDecrypt(b, ConfigFernetTtl, ConfigFernetKeys)
	return string(msg)
}

// Db operations.

func DataDbValidate(db *Db) error {
	err := ValidateSlug("name", db.Name)
	if err != nil {
		return err
	}
	err = ValidatePgUrl("url", db.Url)
	if err != nil {
		return err
	}
	sameNamed, err := DataCount("SELECT count(*) FROM dbs WHERE name=$1 and id!=$2 and removed_at IS NULL", db.Name, db.Id)
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

func DataDbList(queryFrag string) ([]*Db, error) {
	if queryFrag == "" {
		queryFrag = "true"
	}
	res, err := DataConn.Query("SELECT id, name, url_encrypted, added_at, updated_at, version, removed_at FROM dbs WHERE removed_at IS NULL AND " + queryFrag)
	if err != nil {
		return nil, err
	}
	defer func() { Must(res.Close()) }()
	dbs := []*Db{}
	for res.Next() {
		db := Db{}
		urlEncrypted := make([]byte, 0)
		err := res.Scan(&db.Id, &db.Name, &urlEncrypted, &db.AddedAt, &db.UpdatedAt, &db.Version, &db.RemovedAt)
		if err != nil {
			return nil, err
		}
		db.Url = DataFernetDecrypt(urlEncrypted)
		dbs = append(dbs, &db)
	}
	return dbs, nil
}

func DataDbAdd(name string, url string) (*Db, error) {
	db := &Db{
		Id:        uuid.New(),
		Name:      name,
		Url:       url,
		AddedAt:   time.Now(),
		UpdatedAt: time.Now(),
		RemovedAt: nil,
		Version:   1,
	}
	err := DataDbValidate(db)
	if err == nil {
		_, err = DataConn.Exec("INSERT INTO dbs (id, name, url_encrypted, added_at, updated_at, removed_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			db.Id, db.Name, DataFernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Version)
	}
	return db, err
}

func DataDbGet(idOrName string) (*Db, error) {
	var row *sql.Row
	if DataUuidRegexp.MatchString(idOrName) {
		query := "SELECT id, name, url_encrypted, added_at, updated_at, version FROM dbs WHERE removed_at is NULL AND (id=$1 OR name=$2) LIMIT 1"
		row = DataConn.QueryRow(query, idOrName, idOrName)
	} else {
		query := "SELECT id, name, url_encrypted, added_at, updated_at, version FROM dbs WHERE removed_at is NULL AND name=$1 LIMIT 1"
		row = DataConn.QueryRow(query, idOrName)
	}
	db := Db{}
	urlEncrypted := make([]byte, 0)
	err := row.Scan(&db.Id, &db.Name, &urlEncrypted, &db.AddedAt, &db.UpdatedAt, &db.Version)
	switch {
	case err == nil:
		db.Url = DataFernetDecrypt(urlEncrypted)
		return &db, nil
	case err == sql.ErrNoRows:
		return nil, &PgpinError{
			Id:         "db-not-found",
			Message:    "db not found",
			HttpStatus: 404,
		}
	default:
		return nil, err
	}
}

func DataDbUpdate(db *Db) error {
	err := DataDbValidate(db)
	if err != nil {
		return err
	}
	db.UpdatedAt = time.Now()
	result, err := DataConn.Exec("UPDATE dbs SET name=$1, url_encrypted=$2, added_at=$3, updated_at=$4, removed_at=$5, version=$6 WHERE id=$7 AND version=$8",
		db.Name, DataFernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Version+1, db.Id, db.Version)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return &PgpinError{
			Id:         "db-concurrent-update",
			Message:    "concurrent db update attempted",
			HttpStatus: 400,
		}
	}
	db.Version = db.Version + 1
	return nil
}

func DataDbRemove(id string) (*Db, error) {
	db, err := DataDbGet(id)
	if err != nil {
		return nil, err
	}
	numPins, err := DataCount("SELECT count(*) FROM pins WHERE db_id=$1 AND deleted_at IS NULL", db.Id)
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
	err = DataDbUpdate(db)
	return db, err
}

// Pin operations.

func DataPinValidate(pin *Pin) error {
	err := ValidateSlug("name", pin.Name)
	if err != nil {
		return err
	}
	err = ValidateNonempty("query", pin.Query)
	if err != nil {
		return err
	}
	_, err = DataDbGet(pin.DbId)
	if err != nil {
		return err
	}
	sameNamed, err := DataCount("SELECT count(*) FROM pins WHERE name=$1 AND id!=$2 AND deleted_at IS NULL", pin.Name, pin.Id)
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

func DataPinList(queryFrag string) ([]*Pin, error) {
	if queryFrag == "" {
		queryFrag = "true"
	}
	res, err := DataConn.Query("SELECT id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version FROM pins WHERE deleted_at IS NULL AND " + queryFrag)
	if err != nil {
		return nil, err
	}
	defer func() { Must(res.Close()) }()
	pins := []*Pin{}
	for res.Next() {
		pin := Pin{}
		err := res.Scan(&pin.Id, &pin.Name, &pin.DbId, &pin.Query, &pin.CreatedAt, &pin.UpdatedAt, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.ResultsFields, &pin.ResultsRows, &pin.ResultsError, &pin.ScheduledAt, &pin.DeletedAt, &pin.Version)
		if err != nil {
			return nil, err
		}
		pins = append(pins, &pin)
	}
	return pins, nil
}

func DataPinCreate(dbId string, name string, query string) (*Pin, error) {
	now := time.Now()
	pin := &Pin{
		Id:              uuid.New(),
		Name:            name,
		DbId:            dbId,
		Query:           query,
		CreatedAt:       now,
		UpdatedAt:       now,
		QueryStartedAt:  nil,
		QueryFinishedAt: nil,
		ResultsFields:   MustNewPgJson(nil),
		ResultsRows:     MustNewPgJson(nil),
		ResultsError:    nil,
		ScheduledAt:     now,
		DeletedAt:       nil,
		Version:         1,
	}
	err := DataPinValidate(pin)
	if err != nil {
		return nil, err
	}
	_, err = DataConn.Exec("INSERT INTO pins (id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		pin.Id, pin.Name, pin.DbId, pin.Query, pin.CreatedAt, pin.UpdatedAt, pin.QueryStartedAt, pin.QueryFinishedAt, pin.ResultsFields, pin.ResultsRows, pin.ResultsError, pin.ScheduledAt, pin.DeletedAt, pin.Version)
	if err != nil {
		return nil, err
	}
	err = workers.Enqueue("pins", "", pin.Id)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func DataPinGetInternal(queryFrag string, queryVals ...interface{}) (*Pin, error) {
	row := DataConn.QueryRow("SELECT id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version FROM pins WHERE deleted_at IS NULL AND "+queryFrag+" LIMIT 1", queryVals...)
	pin := Pin{}
	err := row.Scan(&pin.Id, &pin.Name, &pin.DbId, &pin.Query, &pin.CreatedAt, &pin.UpdatedAt, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.ResultsFields, &pin.ResultsRows, &pin.ResultsError, &pin.ScheduledAt, &pin.DeletedAt, &pin.Version)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &pin, nil
	}
}

func DataPinGet(idOrName string) (*Pin, error) {
	var pin *Pin
	var err error
	if DataUuidRegexp.MatchString(idOrName) {
		pin, err = DataPinGetInternal("(id=$1 OR name=$2)", idOrName, idOrName)
	} else {
		pin, err = DataPinGetInternal("name=$1", idOrName)
	}
	if err != nil {
		return nil, err
	}
	if pin == nil {
		return nil, &PgpinError{
			Id:         "pin-not-found",
			Message:    "pin not found",
			HttpStatus: 404,
		}
	}
	return pin, nil
}

func DataPinUpdate(pin *Pin) error {
	err := DataPinValidate(pin)
	if err != nil {
		return err
	}
	pin.UpdatedAt = time.Now()
	result, err := DataConn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, updated_at=$5, query_started_at=$6, query_finished_at=$7, results_fields=$8, results_rows=$9, results_error=$10, scheduled_at=$11, deleted_at=$12, version=$13 WHERE id=$14 AND version=$15",
		pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.UpdatedAt, pin.QueryStartedAt, pin.QueryFinishedAt, pin.ResultsFields, pin.ResultsRows, pin.ResultsError, pin.ScheduledAt, pin.DeletedAt, pin.Version+1, pin.Id, pin.Version)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return &PgpinError{
			Id:         "pin-concurrent-update",
			Message:    "concurrent pin updated attempted",
			HttpStatus: 400,
		}
	}
	pin.Version = pin.Version + 1
	return nil
}

func DataPinDelete(id string) (*Pin, error) {
	pin, err := DataPinGet(id)
	if err != nil {
		return nil, err
	}
	deletedAt := time.Now()
	pin.DeletedAt = &deletedAt
	err = DataPinUpdate(pin)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func DataPinDbUrl(pin *Pin) (string, error) {
	db, err := DataDbGet(pin.DbId)
	if err != nil {
		return "", err
	}
	return db.Url, nil
}
