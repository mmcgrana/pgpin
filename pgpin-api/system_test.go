package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/stretchr/testify/assert"
	"github.com/zenazn/goji"
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

func TestStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)
	must(err)
	res := httptest.NewRecorder()
	goji.DefaultMux.ServeHTTP(res, req)
	webStatus(res, req)
	assert.Equal(t, 200, res.Code)
	status := &status{}
	must(json.NewDecoder(res.Body).Decode(status))
	assert.Equal(t, "ok", status.Message)
}

func TestDbAdd(t *testing.T) {
	defer clear()
	in := `{"name": "pins-1", "url": "postgres://u:p@h:1234/d-1"}`
	req, err := http.NewRequest("POST", "/dbs", bytes.NewReader([]byte(in)))
	must(err)
	res := httptest.NewRecorder()
	goji.DefaultMux.ServeHTTP(res, req)
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
}

func TestDbRemove(t *testing.T) {
	defer clear()
}

func TestDbList(t *testing.T) {
	defer clear()
}
