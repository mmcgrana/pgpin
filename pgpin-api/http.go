package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type httpStatusingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *httpStatusingResponseWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

func httpWrapLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		method := req.Method
		path := req.URL.Path
		log("web.request.start method=%s path=%s", method, path)
		wres := httpStatusingResponseWriter{-1, res}
		f(&wres, req)
		elapsed := float64(time.Since(start)) / 1000000.0
		log("web.request.finish method=%s path=%s status=%d elapsed=%f", method, path, wres.status, elapsed)
	}
}

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

func httpReadJson(req *http.Request, reqD interface{}) error {
	return json.NewDecoder(req.Body).Decode(reqD)
}

func httpWriteJson(resp http.ResponseWriter, status int, respD interface{}) {
	b, err := json.MarshalIndent(respD, "", "  ")
	if err != nil {
		panic(err)
	}
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(status)
	resp.Write(b)
	resp.Write([]byte("\n"))
}
