package main

import (
	"code.google.com/p/go-uuid/uuid"
	"database/sql"
	"fmt"
	"github.com/darkhelmet/env"
	fernet "github.com/fernet/fernet-go"
	_ "github.com/lib/pq"
	"log"
	"time"
)

// Constants.

var DataPinResultsRowsMax = 10000
var DataPinRefreshInterval = 10 * time.Minute
var DataPinStatementTimeout = 30 * time.Second
var DataApiStatementTimeout = 5 * time.Second
var DataConnectTimeout = 5 * time.Second
var DataFernetKeys = fernet.MustDecodeKeys(env.String("FERNET_KEY"))
var DataFernetTtl = time.Hour * 24 * 365 * 10

// DB connection.

var DataConn *sql.DB

func DataStart() {
	log.Print("data.start")
	connUrl := fmt.Sprintf("%s?application_name=%s&statement_timeout=%d&connect_timeout=%d",
		env.String("DATABASE_URL"), "pgpin.api", DataApiStatementTimeout / time.Millisecond, DataConnectTimeout / time.Millisecond)
	conn, err := sql.Open("postgres", connUrl)
	if err != nil {
		panic(err)
	}
	conn.SetMaxOpenConns(20)
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
	tok, err := fernet.EncryptAndSign([]byte(s), DataFernetKeys[0])
	Must(err)
	return tok
}

func DataFernetDecrypt(b []byte) string {
	msg := fernet.VerifyAndDecrypt(b, DataFernetTtl, DataFernetKeys)
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

func DataDbList() ([]DbSlim, error) {
	res, err := DataConn.Query("SELECT id, name FROM dbs where removed_at IS NULL")
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

func DataDbAdd(name string, url string) (*Db, error) {
	db := &Db{}
	db.Id = uuid.New()
	db.Name = name
	db.Url = url
	db.AddedAt = time.Now()
	db.UpdatedAt = time.Now()
	err := DataDbValidate(db)
	if err == nil {
		_, err = DataConn.Exec("INSERT INTO dbs (id, name, url_encrypted, added_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			db.Id, db.Name, DataFernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt)
	}
	return db, err
}

func DataDbGet(id string) (*Db, error) {
	row := DataConn.QueryRow(`SELECT id, name, url_encrypted, added_at, updated_at FROM dbs WHERE id=$1 AND removed_at IS NULL LIMIT 1`, id)
	db := Db{}
	urlEncrypted := make([]byte, 0)
	err := row.Scan(&db.Id, &db.Name, &urlEncrypted, &db.AddedAt, &db.UpdatedAt)
	switch {
	case err == nil:
		db.Url = DataFernetDecrypt(urlEncrypted)
		return &db, nil
	case err == sql.ErrNoRows:
		return nil, &PgpinError{
			Id:         "not-found",
			Message:    "db not found",
			HttpStatus: 404,
		}
	default:
		return nil, err
	}
}

func DataDbUpdate(db *Db) error {
	err := DataDbValidate(db)
	if err == nil {
		db.UpdatedAt = time.Now()
		_, err = DataConn.Exec("UPDATE dbs SET name=$1, url_encrypted=$2, added_at=$3, updated_at=$4, removed_at=$5 WHERE id=$6",
			db.Name, DataFernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Id)
	}
	return err
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

func DataPinList() ([]PinSlim, error) {
	res, err := DataConn.Query("SELECT id, name FROM pins where deleted_at IS NULL")
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

func DataPinCreate(dbId string, name string, query string) (*Pin, error) {
	pin := &Pin{}
	pin.Id = uuid.New()
	pin.DbId = dbId
	pin.Name = name
	pin.Query = query
	pin.CreatedAt = time.Now()
	pin.UpdatedAt = time.Now()
	err := DataPinValidate(pin)
	if err != nil {
		return nil, err
	}
	_, err = DataConn.Exec("INSERT INTO pins (id, db_id, name, query, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		pin.Id, pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func DataPinGetInternal(queryFrag string, queryVals ...interface{}) (*Pin, error) {
	row := DataConn.QueryRow(`SELECT id, db_id, name, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, reserved_at FROM pins WHERE deleted_at IS NULL AND `+queryFrag+` LIMIT 1`, queryVals...)
	pin := Pin{}
	err := row.Scan(&pin.Id, &pin.DbId, &pin.Name, &pin.Query, &pin.CreatedAt, &pin.UpdatedAt, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.ResultsFields, &pin.ResultsRows, &pin.ResultsError, &pin.ReservedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &pin, nil
	}
}

func DataPinGet(id string) (*Pin, error) {
	pin, err := DataPinGetInternal("id=$1", id)
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

func DataPinUpdate(pin *Pin) error {
	err := DataPinValidate(pin)
	if err != nil {
		return err
	}
	pin.UpdatedAt = time.Now()
	_, err = DataConn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, updated_at=$5, query_started_at=$6, query_finished_at=$7, results_fields=$8, results_rows=$9, results_error=$10, deleted_at=$11 WHERE id=$12",
		pin.DbId, pin.Name, pin.Query, pin.CreatedAt, pin.UpdatedAt, pin.QueryStartedAt, pin.QueryFinishedAt, pin.ResultsFields, pin.ResultsRows, pin.ResultsError, pin.DeletedAt, pin.Id)
	if err != nil {
		return err
	}
	return nil
}

func DataPinReserve() (*Pin, error) {
	refreshSince := time.Now().Add(-1 * DataPinRefreshInterval)
	pin, err := DataPinGetInternal("((query_started_at is NULL) OR (query_started_at < $1)) AND reserved_at IS NULL AND deleted_at IS NULL", refreshSince)
	if err != nil {
		return nil, err
	}
	if pin == nil {
		return nil, nil
	}
	reservedAt := time.Now()
	pin.ReservedAt = &reservedAt
	err = DataPinUpdate(pin)
	return pin, err
}

func DataPinRelease(pin *Pin) error {
	pin.ReservedAt = nil
	return DataPinUpdate(pin)
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
