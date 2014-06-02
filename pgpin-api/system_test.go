package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	if !strings.HasSuffix(env.String("DATABASE_URL"), "-test") {
		panic("Doesn't look like a test database")
	}
	dataStart()
}

func clear() {
	_, err := dataConn.Exec("DELETE from dbs")
	must(err)
	_, err = dataConn.Exec("DELETE from pins")
	must(err)
}

func TestStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "https://pgpin.com/status", nil)
	must(err)
	res := httptest.NewRecorder()
	webStatus(res, req)
	if res.Code != 200 {
		t.Errorf("Got status %d, want 200", res.Code)
	}
	s := &status{}
	err = json.NewDecoder(res.Body).Decode(s)
	if s.Message != "ok" {
		t.Errorf("Got message %s, want ok")
	}
}

func TestAddDb(t *testing.T) {
	defer clear()
	in := `{"name": "pins", "url": "postgres://u:p@h:1234/d"}`
	req, err := http.NewRequest("GET", "http://pgpin.com/status", bytes.NewReader([]byte(in)))
	must(err)
	res := httptest.NewRecorder()
	webDbAdd(res, req)
	if !(res.Code == 201) {
		t.Errorf("Got status %d, want 201", res.Code)
	}
	db := &db{}
	must(json.NewDecoder(res.Body).Decode(db))
	if !(db.Name == "pins") {
		t.Errorf("Got name %s, want pins", db.Name)
	}
	if !(db.Url == "postgres://u:p@h:1234/d") {
		t.Errorf("Wrong Url %s", db.Url)
	}
	if !(len(db.Id) == 12) {
		t.Errorf("Expeced 12-char Id, got %s", db.Id)
	}
	if !(db.AddedAt.After(time.Now().Add(-5*time.Second))) {
		t.Errorf("Expected recent time, got %v", db.AddedAt)
	}
}
