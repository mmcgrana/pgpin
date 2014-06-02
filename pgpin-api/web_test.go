package main

import (
	"encoding/json"
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
	os.Setenv("DATABASE_URL", os.Getenv("TEST_DATABASE_URL"))
	DataStart()
}

func TestStatusOk(t *testing.T) {
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
