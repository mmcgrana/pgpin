package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Middleware.

func TestRequestIdGenerated(t *testing.T) {
	res := mustRequest("GET", "/status", nil)
	assert.True(t, DataUuidRegexp.MatchString(res.Header().Get("Request-Id")))
}

func TestRequestIdGiven(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)
	Must(err)
	req.Header.Set("Request-Id", "given")
	res := httptest.NewRecorder()
	WebMux.ServeHTTP(res, req)
	assert.Equal(t, "given", res.Header().Get("Request-Id"))
}

// DB endpoints.

func TestDbCreate(t *testing.T) {
	defer clear()
	b := asReader(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-1"}`)
	res := mustRequest("POST", "/v1/dbs", b)
	assert.Equal(t, 201, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.NotEmpty(t, dbOut.Id)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbCreateDuplicateName(t *testing.T) {
	defer clear()
	mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	b := asReader(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-other"}`)
	res := mustRequest("POST", "/v1/dbs", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-db-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbGet(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("GET", "/v1/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), dbOut.UpdatedAt, 3*time.Second)
}

func TestDbGetByName(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("GET", "/v1/dbs/dbs-1", nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "dbs-1", dbOut.Name)
}

func TestDbUpdateName(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	b := asReader(`{"name": "dbs-1a"}`)
	res := mustRequest("PUT", "/v1/dbs/"+dbIn.Id, b)
	assert.Equal(t, 200, res.Code)
	dbPutOut := &Db{}
	mustDecode(res, dbPutOut)
	assert.Equal(t, "dbs-1a", dbPutOut.Name)
	res = mustRequest("GET", "/v1/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbGetOut := &Db{}
	mustDecode(res, dbGetOut)
	assert.Equal(t, "dbs-1a", dbGetOut.Name)
	assert.True(t, dbGetOut.UpdatedAt.After(dbIn.UpdatedAt))
}

func TestDbUpdateUrl(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	b := asReader(`{"url": "postgres://u:p@h:1234/d-1a"}`)
	res := mustRequest("PUT", "/v1/dbs/"+dbIn.Id, b)
	assert.Equal(t, 200, res.Code)
	dbPutOut := &Db{}
	mustDecode(res, dbPutOut)
	assert.Equal(t, "dbs-1", dbPutOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1a", dbPutOut.Url)
	res = mustRequest("GET", "/v1/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbGetOut := &Db{}
	mustDecode(res, dbGetOut)
	assert.Equal(t, "dbs-1", dbPutOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1a", dbPutOut.Url)
	assert.True(t, dbGetOut.UpdatedAt.After(dbIn.UpdatedAt))
}

func TestDbDelete(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("DELETE", "/v1/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	res = mustRequest("GET", "/v1/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestDbDeleteWithPins(t *testing.T) {
	defer clear()
	db := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	mustPinCreate(db.Id, "pins-1", "select count (*) from pins")
	res := mustRequest("DELETE", "/v1/dbs/"+db.Id, nil)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "removing-db-with-pins", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbList(t *testing.T) {
	defer clear()
	dbIn1 := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	dbIn2 := mustDbCreate("dbs-2", "postgres://u:p@h:1234/d-2")
	_, err := DbDelete(dbIn2.Id)
	Must(err)
	res := mustRequest("GET", "/v1/dbs", nil)
	assert.Equal(t, 200, res.Code)
	dbsOut := []*Db{}
	mustDecode(res, &dbsOut)
	assert.Equal(t, len(dbsOut), 1)
	assert.Equal(t, dbIn1.Id, dbsOut[0].Id)
	assert.Equal(t, "dbs-1", dbsOut[0].Name)
}

// Pin endpoints.

func TestPinCreateAndGet(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	b := asReader(`{"name": "pin-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`)
	res := mustRequest("POST", "/v1/pins", b)
	assert.Equal(t, 201, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	mustWorkerTick()
	res = mustRequest("GET", "/v1/pins/"+pinOut.Id, nil)
	assert.Equal(t, 200, res.Code)
	mustDecode(res, pinOut)
	assert.NotEmpty(t, pinOut.Id)
	assert.Equal(t, "pin-1", pinOut.Name)
	assert.Equal(t, dbIn.Id, pinOut.DbId)
	assert.Equal(t, "select count(*) from pins", pinOut.Query)
	assert.WithinDuration(t, time.Now(), pinOut.CreatedAt, 3*time.Second)
	assert.True(t, pinOut.QueryStartedAt.After(pinOut.CreatedAt))
	assert.True(t, pinOut.QueryFinishedAt.After(*pinOut.QueryStartedAt))
	assert.Equal(t, `["count"]`, mustCanonicalizeJson(pinOut.ResultsFields))
	assert.Equal(t, `[[1]]`, mustCanonicalizeJson(pinOut.ResultsRows))
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinGetByName(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	res := mustRequest("GET", "/v1/pins/pins-1", nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, pinIn.Id, pinOut.Id)
	assert.Equal(t, "pins-1", pinOut.Name)
}

func TestPinCreateDuplicateName(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := asReader(`{"name": "pins-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`)
	res := mustRequest("POST", "/v1/pins", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-pin-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestPinUpdateName(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := asReader(`{"name": "pins-1a"}`)
	res := mustRequest("PUT", "/v1/pins/"+pinIn.Id, b)
	assert.Equal(t, 200, res.Code)
	pinPutOut := &Pin{}
	mustDecode(res, pinPutOut)
	assert.Equal(t, "pins-1a", pinPutOut.Name)
	res = mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinGetOut := &Pin{}
	mustDecode(res, pinGetOut)
	assert.Equal(t, "pins-1a", pinGetOut.Name)
	assert.True(t, pinGetOut.UpdatedAt.After(pinIn.UpdatedAt))
}

func TestPinUpdateQuery(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", "postgres://u:p@h:1234/d-1")
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := asReader(`{"query": "select now()"}`)
	res := mustRequest("PUT", "/v1/pins/"+pinIn.Id, b)
	assert.Equal(t, 200, res.Code)
	pinPutOut := &Pin{}
	mustDecode(res, pinPutOut)
	assert.Equal(t, "select now()", pinPutOut.Query)
	res = mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinGetOut := &Pin{}
	mustDecode(res, pinGetOut)
	assert.Equal(t, "pins-1", pinGetOut.Name)
	assert.Equal(t, "select now()", pinGetOut.Query)
	assert.True(t, pinGetOut.UpdatedAt.After(pinIn.UpdatedAt))
}

func TestPinMultipleColumns(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select name, query from pins")
	mustWorkerTick()
	res := mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, `["name","query"]`, mustCanonicalizeJson(pinOut.ResultsFields))
	assert.Equal(t, `[["pins-1","select name, query from pins"]]`, mustCanonicalizeJson(pinOut.ResultsRows))
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinTooManyRows(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select generate_series(0, 10000)")
	mustWorkerTick()
	res := mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "too many rows in query results", *pinOut.ResultsError)
}

func TestPinBadQuery(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select wat")
	mustWorkerTick()
	res := mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "column \"wat\" does not exist", *pinOut.ResultsError)
}

func TestPinStatementTimeout(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select pg_sleep(0.1)")
	ConfigPinStatementTimeoutPrev := ConfigPinStatementTimeout
	defer func() {
		ConfigPinStatementTimeout = ConfigPinStatementTimeoutPrev
	}()
	ConfigPinStatementTimeout = time.Millisecond * 50
	mustWorkerTick()
	res := mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "canceling statement due to statement timeout", *pinOut.ResultsError)
}

func TestPinMalformedDbUrl(t *testing.T) {
	defer clear()
	_, err := DbCreate("dbs-1", "not-a-url")
	assert.Equal(t, "pgpin: invalid: field url must be a valid postgres:// URL", err.Error())
}

func TestPinUnreachableRbUrl(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl+"-moar")
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	mustWorkerTick()
	res := mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "could not connect to database", *pinOut.ResultsError)
}

func TestPinOptomisticLocking(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinWinsRace := mustPinCreate(dbIn.Id, "pins-1", "select 1")
	pinLosesRace := mustPinGet(pinWinsRace.Id)
	pinWinsRace.Query = "select 'wins'"
	err := PinUpdate(pinWinsRace)
	assert.Nil(t, err)
	pinLosesRace.Query = "select 'loses'"
	err = PinUpdate(pinLosesRace)
	assert.Equal(t, "pin-concurrent-update", err.(*PgpinError).Id)
	pinAfterRace := mustPinGet(pinWinsRace.Id)
	assert.Equal(t, "select 'wins'", pinAfterRace.Query)
}

func TestPinDelete(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	res := mustRequest("DELETE", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "pins-1", pinOut.Name)
	res = mustRequest("GET", "/v1/pins/"+pinIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestPinList(t *testing.T) {
	defer clear()
	dbIn := mustDbCreate("dbs-1", ConfigDatabaseUrl)
	pinIn1 := mustPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	pinIn2 := mustPinCreate(dbIn.Id, "pins-2", "select * from pins")
	_, err := PinDelete(pinIn1.Id)
	Must(err)
	res := mustRequest("GET", "/v1/pins", nil)
	assert.Equal(t, 200, res.Code)
	pinsOut := []*Pin{}
	mustDecode(res, &pinsOut)
	assert.Equal(t, len(pinsOut), 1)
	assert.Equal(t, pinIn2.Id, pinsOut[0].Id)
	assert.Equal(t, "pins-2", pinsOut[0].Name)
}

// Misc endpoints.

func TestStatus(t *testing.T) {
	res := mustRequest("GET", "/status", nil)
	assert.Equal(t, 200, res.Code)
	status := &Status{}
	mustDecode(res, status)
	assert.Equal(t, "ok", status.Message)
}

func TestError(t *testing.T) {
	res := mustRequest("GET", "/error", nil)
	assert.Equal(t, 500, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "internal-error", data["id"])
	assert.Equal(t, "internal server error", data["message"])
}

func TestPanic(t *testing.T) {
	res := mustRequest("GET", "/panic", nil)
	assert.Equal(t, 500, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "internal-error", data["id"])
	assert.Equal(t, "internal server error", data["message"])
}

func TestTimeout(t *testing.T) {
	WebMuxPrev := WebMux
	defer func() {
		WebMux = WebMuxPrev
	}()
	ConfigWebTimeoutPrev := ConfigWebTimeout
	defer func() {
		ConfigWebTimeout = ConfigWebTimeoutPrev
	}()
	ConfigWebTimeout = 50 * time.Millisecond
	WebBuild()
	res := mustRequest("GET", "/timeout", nil)
	assert.Equal(t, 503, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "request-timeout", data["id"])
	assert.Equal(t, "request timed out", data["message"])
}

func TestNotFound(t *testing.T) {
	res := mustRequest("GET", "/wat", nil)
	assert.Equal(t, 404, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "not-found", data["id"])
	assert.Equal(t, "not found", data["message"])
}
