package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/stretchr/testify/assert"
	"github.com/zenazn/goji"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	log.SetOutput(ioutil.Discard)
	if !strings.HasSuffix(env.String("DATABASE_URL"), "-test") {
		panic("Doesn't look like a test database")
	}
	dataStart()
	webBuild()
}

func clear() {
	_, err := dataConn.Exec("DELETE from dbs")
	must(err)
	_, err = dataConn.Exec("DELETE from pins")
	must(err)
}

func mustRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	must(err)
	res := httptest.NewRecorder()
	goji.DefaultMux.ServeHTTP(res, req)
	return res
}

func TestStatus(t *testing.T) {
	res := mustRequest("GET", "/status", nil)
	assert.Equal(t, 200, res.Code)
	status := &status{}
	must(json.NewDecoder(res.Body).Decode(status))
	assert.Equal(t, "ok", status.Message)
}

func TestDbAdd(t *testing.T) {
	defer clear()
	b := bytes.NewReader([]byte(`{"name": "pins-1", "url": "postgres://u:p@h:1234/d-1"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 201, res.Code)
	db := &db{}
	must(json.NewDecoder(res.Body).Decode(db))
	assert.Equal(t, "pins-1", db.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", db.Url)
	assert.NotEmpty(t, db.Id)
	assert.WithinDuration(t, time.Now(), db.AddedAt, 3*time.Second)
}

func TestDbGet(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("pins-1", "postgres://u:p@h:1234/d-1")
	must(err)
	res := mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &db{}
	must(json.NewDecoder(res.Body).Decode(dbOut))
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "pins-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbRemove(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("pins-1", "postgres://u:p@h:1234/d-1")
	must(err)
	res := mustRequest("DELETE", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestDbListBasic(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("pins-1", "postgres://u:p@h:1234/d-1")
	must(err)
	res := mustRequest("GET", "/dbs", nil)
	assert.Equal(t, 200, res.Code)
	dbsOut := []*db{}
	must(json.NewDecoder(res.Body).Decode(&dbsOut))
	assert.Equal(t, len(dbsOut), 1)
	assert.Equal(t, dbIn.Id, dbsOut[0].Id)
	assert.Equal(t, "pins-1", dbsOut[0].Name)
}

func TestDBListDeletions(t *testing.T) {
	defer clear()
	dbIn1, err := dataDbAdd("pins-1", "postgres://u:p@h:1234/d-1")
	must(err)
	dbIn2, err := dataDbAdd("pins-2", "postgres://u:p@h:1234/d-2")
	must(err)
	_, err = dataDbRemove(dbIn2.Id)
	must(err)
	res := mustRequest("GET", "/dbs", nil)
	assert.Equal(t, 200, res.Code)
	dbsOut := []*db{}
	must(json.NewDecoder(res.Body).Decode(&dbsOut))
	assert.Equal(t, len(dbsOut), 1)
	assert.Equal(t, dbIn1.Id, dbsOut[0].Id)
}

func TestPinCreate(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("pins-1", "postgres://u:p@h:1234/d-1")
	must(err)
	b := bytes.NewReader([]byte(`{"name": "pin-1", "db_id": "` + dbIn.Id + `", "query": "select * from pins"`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 201, res.Code)
	workerTick()
	pinOut := &pin{}
	must(json.NewDecoder(res.Body).Decode(pinOut))
	assert.Equal(t, "pin-1", pinOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", db.Url)
	assert.NotEmpty(t, db.Id)
	assert.WithinDuration(t, time.Now(), db.AddedAt, 3*time.Second)
}

func TestPinWithoutDb(t *testing.T) {
	defer clear()
}
