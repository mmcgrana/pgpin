package main

import (
	"encoding/json"
	"env"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func must(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func init() {
	if !strings.HasSuffix(env.String("DATABASE_URL"), "-test") {
		panic("Doesn't look like a test database")
	}
	dataStart()
	dataConn
}

func TestStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "https://pgpin-api.com/status", nil)
	must(t, err)
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
