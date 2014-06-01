package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

type authenticator func(string, string) bool

func getAuth(r *http.Request) (string, string, bool) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return "", "", false
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return "", "", false
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return "", "", false
	}
	return pair[0], pair[1], true
}
