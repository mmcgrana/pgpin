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

func clear() {
	_, err := dataConn.Exec("DELETE from pins")
	must(err)
	_, err = dataConn.Exec("DELETE from dbs")
	must(err)
}

func init() {
	log.SetOutput(ioutil.Discard)
	if !strings.HasSuffix(env.String("DATABASE_URL"), "-test") {
		panic("Doesn't look like a test database")
	}
	dataStart()
	clear()
	webBuild()
}

func mustRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	must(err)
	res := httptest.NewRecorder()
	goji.DefaultMux.ServeHTTP(res, req)
	return res
}

func mustDecode(res *httptest.ResponseRecorder, data interface{}) {
	must(json.NewDecoder(res.Body).Decode(data))
}

func TestNotFound(t *testing.T) {
	res := mustRequest("GET", "/wat", nil)
	assert.Equal(t, 404, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "not-found", data["id"])
	assert.Equal(t, "not found", data["message"])
}

func TestStatus(t *testing.T) {
	res := mustRequest("GET", "/status", nil)
	assert.Equal(t, 200, res.Code)
	status := &status{}
	mustDecode(res, status)
	assert.Equal(t, "ok", status.Message)
}

func TestDbAdd(t *testing.T) {
	defer clear()
	b := bytes.NewReader([]byte(`{"name": "pins-1", "url": "postgres://u:p@h:1234/d-1"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 201, res.Code)
	dbOut := &db{}
	mustDecode(res, dbOut)
	assert.Equal(t, "pins-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.NotEmpty(t, dbOut.Id)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbGet(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	must(err)
	res := mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &db{}
	mustDecode(res, dbOut)
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbRemove(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	must(err)
	res := mustRequest("DELETE", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestDBList(t *testing.T) {
	defer clear()
	dbIn1, err := dataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	must(err)
	dbIn2, err := dataDbAdd("dbs-2", "postgres://u:p@h:1234/d-2")
	must(err)
	_, err = dataDbRemove(dbIn2.Id)
	must(err)
	res := mustRequest("GET", "/dbs", nil)
	assert.Equal(t, 200, res.Code)
	dbsOut := []*db{}
	mustDecode(res, &dbsOut)
	assert.Equal(t, len(dbsOut), 1)
	assert.Equal(t, dbIn1.Id, dbsOut[0].Id)
	assert.Equal(t, "dbs-1", dbsOut[0].Name)
}

func TestPinCreateAndGet(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("dbs-1", env.String("DATABASE_URL"))
	must(err)
	b := bytes.NewReader([]byte(`{"name": "pin-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`))
	res := mustRequest("POST", "/pins", b)
	assert.Equal(t, 201, res.Code)
	pinOut := &pin{}
	mustDecode(res, pinOut)
	workerTick()
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
	assert.Equal(t, `["count"]`, *pinOut.ResultsFieldsJson)
	assert.Equal(t, `[[1]]`, *pinOut.ResultsRowsJson)
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinDelete(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("dbs-1", env.String("DATABASE_URL"))
	must(err)
	pinIn, err := dataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	must(err)
	res := mustRequest("DELETE", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "pins-1", pinOut.Name)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestPinList(t *testing.T) {
	defer clear()
	dbIn, err := dataDbAdd("dbs-1", env.String("DATABASE_URL"))
	must(err)
	pinIn1, err := dataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	must(err)
	pinIn2, err := dataPinCreate(dbIn.Id, "pins-2", "select * from pins")
	must(err)
	_, err = dataPinDelete(pinIn1.Id)
	must(err)
	res := mustRequest("GET", "/pins", nil)
	assert.Equal(t, 200, res.Code)
	pinsOut := []*pin{}
	mustDecode(res, &pinsOut)
	assert.Equal(t, len(pinsOut), 1)
	assert.Equal(t, pinIn2.Id, pinsOut[0].Id)
	assert.Equal(t, "pins-2", pinsOut[0].Name)
}
