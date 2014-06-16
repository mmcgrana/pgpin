package main

import (
	"code.google.com/p/go-uuid/uuid"
	"database/sql"
	"github.com/jrallison/go-workers"
	_ "github.com/lib/pq"
	"regexp"
	"time"
)

// Constants.

var DataUuidRegexp = regexp.MustCompilePOSIX("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")

// Structs.

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
	ScheduledAt     time.Time  `json:"-"`
	DeletedAt       *time.Time `json:"-"`
	Version         int        `json:"-"`
}

type Db struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	AddedAt   time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RemovedAt *time.Time `json:"-"`
	Version   int        `json:"-"`
}

// Db CRUD.

func DbValidate(db *Db) error {
	err := ValidateSlug("name", db.Name)
	if err != nil {
		return err
	}
	err = ValidatePgUrl("url", db.Url)
	if err != nil {
		return err
	}
	sameNamed, err := PgCount("SELECT count(*) FROM dbs WHERE name=$1 and id!=$2 and deleted_at IS NULL", db.Name, db.Id)
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

func DbList(queryFrag string) ([]*Db, error) {
	if queryFrag == "" {
		queryFrag = "true"
	}
	res, err := PgConn.Query("SELECT id, name, url_encrypted, created_at, updated_at, version, deleted_at FROM dbs WHERE deleted_at IS NULL AND " + queryFrag)
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
		db.Url = FernetDecrypt(urlEncrypted)
		dbs = append(dbs, &db)
	}
	return dbs, nil
}

func DbCreate(name string, url string) (*Db, error) {
	db := &Db{
		Id:        uuid.New(),
		Name:      name,
		Url:       url,
		AddedAt:   time.Now(),
		UpdatedAt: time.Now(),
		RemovedAt: nil,
		Version:   1,
	}
	err := DbValidate(db)
	if err == nil {
		_, err = PgConn.Exec("INSERT INTO dbs (id, name, url_encrypted, created_at, updated_at, deleted_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			db.Id, db.Name, FernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Version)
	}
	return db, err
}

func DbGet(idOrName string) (*Db, error) {
	var row *sql.Row
	if DataUuidRegexp.MatchString(idOrName) {
		query := "SELECT id, name, url_encrypted, created_at, updated_at, version FROM dbs WHERE deleted_at is NULL AND (id=$1 OR name=$2) LIMIT 1"
		row = PgConn.QueryRow(query, idOrName, idOrName)
	} else {
		query := "SELECT id, name, url_encrypted, created_at, updated_at, version FROM dbs WHERE deleted_at is NULL AND name=$1 LIMIT 1"
		row = PgConn.QueryRow(query, idOrName)
	}
	db := Db{}
	urlEncrypted := make([]byte, 0)
	err := row.Scan(&db.Id, &db.Name, &urlEncrypted, &db.AddedAt, &db.UpdatedAt, &db.Version)
	switch {
	case err == nil:
		db.Url = FernetDecrypt(urlEncrypted)
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

func DbUpdate(db *Db) error {
	err := DbValidate(db)
	if err != nil {
		return err
	}
	db.UpdatedAt = time.Now()
	result, err := PgConn.Exec("UPDATE dbs SET name=$1, url_encrypted=$2, created_at=$3, updated_at=$4, deleted_at=$5, version=$6 WHERE id=$7 AND version=$8",
		db.Name, FernetEncrypt(db.Url), db.AddedAt, db.UpdatedAt, db.RemovedAt, db.Version+1, db.Id, db.Version)
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

func DbDelete(id string) (*Db, error) {
	db, err := DbGet(id)
	if err != nil {
		return nil, err
	}
	numPins, err := PgCount("SELECT count(*) FROM pins WHERE db_id=$1 AND deleted_at IS NULL", db.Id)
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
	err = DbUpdate(db)
	return db, err
}

// Pin operations.

func PinValidate(pin *Pin) error {
	err := ValidateSlug("name", pin.Name)
	if err != nil {
		return err
	}
	err = ValidateNonempty("query", pin.Query)
	if err != nil {
		return err
	}
	_, err = DbGet(pin.DbId)
	if err != nil {
		return err
	}
	sameNamed, err := PgCount("SELECT count(*) FROM pins WHERE name=$1 AND id!=$2 AND deleted_at IS NULL", pin.Name, pin.Id)
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

func PinList(queryFrag string, queryVals ...interface{}) ([]*Pin, error) {
	if queryFrag == "" {
		queryFrag = "true"
	}
	query := "SELECT id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version FROM pins WHERE deleted_at IS NULL AND " + queryFrag
	res, err := PgConn.Query(query, queryVals...)
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

func PinCreate(dbId string, name string, query string) (*Pin, error) {
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
	err := PinValidate(pin)
	if err != nil {
		return nil, err
	}
	_, err = PgConn.Exec("INSERT INTO pins (id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
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

func PinGetInternal(queryFrag string, queryVals ...interface{}) (*Pin, error) {
	row := PgConn.QueryRow("SELECT id, name, db_id, query, created_at, updated_at, query_started_at, query_finished_at, results_fields, results_rows, results_error, scheduled_at, deleted_at, version FROM pins WHERE deleted_at IS NULL AND "+queryFrag+" LIMIT 1", queryVals...)
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

func PinGet(idOrName string) (*Pin, error) {
	var pin *Pin
	var err error
	if DataUuidRegexp.MatchString(idOrName) {
		pin, err = PinGetInternal("(id=$1 OR name=$2)", idOrName, idOrName)
	} else {
		pin, err = PinGetInternal("name=$1", idOrName)
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

func PinUpdate(pin *Pin) error {
	err := PinValidate(pin)
	if err != nil {
		return err
	}
	pin.UpdatedAt = time.Now()
	result, err := PgConn.Exec("UPDATE pins SET db_id=$1, name=$2, query=$3, created_at=$4, updated_at=$5, query_started_at=$6, query_finished_at=$7, results_fields=$8, results_rows=$9, results_error=$10, scheduled_at=$11, deleted_at=$12, version=$13 WHERE id=$14 AND version=$15",
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

func PinDelete(id string) (*Pin, error) {
	pin, err := PinGet(id)
	if err != nil {
		return nil, err
	}
	deletedAt := time.Now()
	pin.DeletedAt = &deletedAt
	err = PinUpdate(pin)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func PinDbUrl(pin *Pin) (string, error) {
	db, err := DbGet(pin.DbId)
	if err != nil {
		return "", err
	}
	return db.Url, nil
}
