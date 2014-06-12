package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Setup and teardown.

func init() {
	if env.StringDefault("TEST_LOGS", "false") != "true" {
		log.SetOutput(ioutil.Discard)
	}
	os.Setenv("DATABASE_URL", env.String("TEST_DATABASE_URL"))
	DataStart()
	clear()
	WebBuild()
}

func clear() {
	_, err := DataConn.Exec("DELETE from pins")
	Must(err)
	_, err = DataConn.Exec("DELETE from dbs")
	Must(err)
}

// Helpers.

func mustRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	Must(err)
	res := httptest.NewRecorder()
	WebMux.ServeHTTP(res, req)
	return res
}

func mustDecode(res *httptest.ResponseRecorder, data interface{}) {
	Must(json.NewDecoder(res.Body).Decode(data))
}

func mustDataDbAdd(name string, url string) *Db {
	db, err := DataDbAdd(name, url)
	Must(err)
	return db
}

func mustDataPinCreate(dbId string, name string, query string) *Pin {
	pin, err := DataPinCreate(dbId, name, query)
	Must(err)
	return pin
}

func mustDataPinGet(id string) *Pin {
	pin, err := DataPinGet(id)
	Must(err)
	return pin
}

func mustWorkerTick() {
	_, err := WorkerTick()
	Must(err)
}

func mustCanonicalizeJson(in []byte) string {
	data := make([]interface{}, 0)
	Must(json.Unmarshal(in, &data))
	out, err := json.Marshal(data)
	Must(err)
	return string(out)

}

// DB endpoints.

func TestDbAdd(t *testing.T) {
	defer clear()
	b := bytes.NewReader([]byte(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-1"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 201, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.NotEmpty(t, dbOut.Id)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbAddDuplicateName(t *testing.T) {
	defer clear()
	mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	b := bytes.NewReader([]byte(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-other"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-db-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbGet(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), dbOut.UpdatedAt, 3*time.Second)
}

func TestDbUpdateName(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	b := bytes.NewReader([]byte(`{"name": "dbs-1a"}`))
	res := mustRequest("PUT", "/dbs/"+dbIn.Id, b)
	assert.Equal(t, 200, res.Code)
	dbPutOut := &Db{}
	mustDecode(res, dbPutOut)
	assert.Equal(t, "dbs-1a", dbPutOut.Name)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbGetOut := &Db{}
	mustDecode(res, dbGetOut)
	assert.Equal(t, "dbs-1a", dbGetOut.Name)
	assert.True(t, dbGetOut.UpdatedAt.After(dbIn.UpdatedAt))
}

func TestDbUpdateUrl(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	b := bytes.NewReader([]byte(`{"url": "postgres://u:p@h:1234/d-1a"}`))
	res := mustRequest("PUT", "/dbs/"+dbIn.Id, b)
	assert.Equal(t, 200, res.Code)
	dbPutOut := &Db{}
	mustDecode(res, dbPutOut)
	assert.Equal(t, "dbs-1", dbPutOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1a", dbPutOut.Url)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbGetOut := &Db{}
	mustDecode(res, dbGetOut)
	assert.Equal(t, "dbs-1", dbPutOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1a", dbPutOut.Url)
	assert.True(t, dbGetOut.UpdatedAt.After(dbIn.UpdatedAt))
}

func TestDbRemove(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("DELETE", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestDbRemoveWithPins(t *testing.T) {
	defer clear()
	db := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	mustDataPinCreate(db.Id, "pins-1", "select count (*) from pins")
	res := mustRequest("DELETE", "/dbs/"+db.Id, nil)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "removing-db-with-pins", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbList(t *testing.T) {
	defer clear()
	dbIn1 := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	dbIn2 := mustDataDbAdd("dbs-2", "postgres://u:p@h:1234/d-2")
	_, err := DataDbRemove(dbIn2.Id)
	Must(err)
	res := mustRequest("GET", "/dbs", nil)
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
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	b := bytes.NewReader([]byte(`{"name": "pin-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`))
	res := mustRequest("POST", "/pins", b)
	assert.Equal(t, 201, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	mustWorkerTick()
	res = mustRequest("GET", "/pins/"+pinOut.Id, nil)
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

func TestPinCreateDuplicateName(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := bytes.NewReader([]byte(`{"name": "pins-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`))
	res := mustRequest("POST", "/pins", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-pin-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestPinUpdateName(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := bytes.NewReader([]byte(`{"name": "pins-1a"}`))
	res := mustRequest("PUT", "/pins/"+pinIn.Id, b)
	assert.Equal(t, 200, res.Code)
	pinPutOut := &Pin{}
	mustDecode(res, pinPutOut)
	assert.Equal(t, "pins-1a", pinPutOut.Name)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinGetOut := &Pin{}
	mustDecode(res, pinGetOut)
	assert.Equal(t, "pins-1a", pinGetOut.Name)
	assert.True(t, pinGetOut.UpdatedAt.After(pinIn.UpdatedAt))
}

func TestPinUpdateQuery(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := bytes.NewReader([]byte(`{"query": "select now()"}`))
	res := mustRequest("PUT", "/pins/"+pinIn.Id, b)
	assert.Equal(t, 200, res.Code)
	pinPutOut := &Pin{}
	mustDecode(res, pinPutOut)
	assert.Equal(t, "select now()", pinPutOut.Query)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinGetOut := &Pin{}
	mustDecode(res, pinGetOut)
	assert.Equal(t, "pins-1", pinGetOut.Name)
	assert.Equal(t, "select now()", pinGetOut.Query)
	assert.True(t, pinGetOut.UpdatedAt.After(pinIn.UpdatedAt))
}

func TestPinMultipleColumns(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select name, query from pins")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, `["name","query"]`, mustCanonicalizeJson(pinOut.ResultsFields))
	assert.Equal(t, `[["pins-1","select name, query from pins"]]`, mustCanonicalizeJson(pinOut.ResultsRows))
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinTooManyRows(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select generate_series(0, 10000)")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "too many rows in query results", *pinOut.ResultsError)
}

func TestPinBadQuery(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select wat")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "column \"wat\" does not exist", *pinOut.ResultsError)
}

func TestPinBadDbUrl(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL")+"-moar")
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "could not connect to database", *pinOut.ResultsError)
}

func TestPinDelete(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	res := mustRequest("DELETE", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "pins-1", pinOut.Name)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestPinList(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn1 := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	pinIn2 := mustDataPinCreate(dbIn.Id, "pins-2", "select * from pins")
	_, err := DataPinDelete(pinIn1.Id)
	Must(err)
	res := mustRequest("GET", "/pins", nil)
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

func TestNotFound(t *testing.T) {
	res := mustRequest("GET", "/wat", nil)
	assert.Equal(t, 404, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "not-found", data["id"])
	assert.Equal(t, "not found", data["message"])
}

// Worker behaviour.

func TestWorkerNoop(t *testing.T) {
	processed, err := WorkerTick()
	assert.False(t, processed)
	assert.Nil(t, err)
}

func TestWorkerPinRefreshNotReady(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustDataPinGet(pinIn.Id)
	mustWorkerTick()
	pinOut2 := mustDataPinGet(pinIn.Id)
	assert.Equal(t, string([]byte(pinOut1.ResultsRows)), string([]byte(pinOut2.ResultsRows)))
}

func TestWorkerPinRefreshReady(t *testing.T) {
	defer clear()
	DataPinRefreshIntervalPrev := DataPinRefreshInterval
	defer func() {
		DataPinRefreshInterval = DataPinRefreshIntervalPrev
	}()
	DataPinRefreshInterval = 0 * time.Second
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustDataPinGet(pinIn.Id)
	mustWorkerTick()
	pinOut2 := mustDataPinGet(pinIn.Id)
	assert.NotEqual(t, string([]byte(pinOut1.ResultsRows)), string([]byte(pinOut2.ResultsRows)))
}
