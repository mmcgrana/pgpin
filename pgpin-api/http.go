package main

import (
	"code.google.com/p/gorilla/mux"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type ResponseWriterFlusher interface {
	http.ResponseWriter
	http.Flusher
}

type statusCapturingResponseWriter struct {
	status int
	ResponseWriterFlusher
}

func (w *statusCapturingResponseWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriterFlusher.WriteHeader(s)
}

func newStatusCapturingResponseWriter(w http.ResponseWriter) statusCapturingResponseWriter {
	wf, ok := w.(ResponseWriterFlusher)
	if !ok {
		panic("unflushable")
	}
	return statusCapturingResponseWriter{-1, wf}
}

func wrapLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		method := req.Method
		path := req.URL.Path
		log("key=web.request.start method=%s path=%s", method, path)
		wres := newStatusCapturingResponseWriter(res)
		f(&wres, req)
		elapsed := float64(time.Since(start)) / 1000000.0
		log("key=web.request.finish method=%s path=%s status=%d elapsed=%f", method, path, wres.status, elapsed)
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

func readJson(resp http.ResponseWriter, req *http.Request, reqD interface{}) bool {
	err := json.NewDecoder(req.Body).Decode(reqD)
	if err != nil {
		return false
	}
	return true
}

func writeJson(resp http.ResponseWriter, status int, respD interface{}) {
	b, err := json.MarshalIndent(respD, "", "  ")
	if err != nil {
		panic(err)
	}
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(status)
	resp.Write(b)
	resp.Write([]byte("\n"))
}

func param(req *http.Request, name string) string {
	s := mux.Vars(req)[name]
	if s != "" {
		return s
	}
	return req.FormValue(name)
}

func routerHandlerFunc(router *mux.Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		router.ServeHTTP(res, req)
	}
}

func wrapReqCount(h http.HandlerFunc, reqCountPtr *int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(reqCountPtr, 1)
		h(w, r)
		atomic.AddInt64(reqCountPtr, -1)
	}
}


func httpServeGraceful(handler http.HandlerFunc, port int, sigs chan os.Signal) {
	var reqCount int64 = 0
	handler = wrapReqCount(handler, &reqCount)
	server := &http.Server{Handler: handler}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	stop := make(chan bool, 1)
	go func() {
		<-sigs
		err = listener.Close()
		if err != nil {
			panic(err)
		}
		for {
			reqCountCurrent := atomic.LoadInt64(&reqCount)
			if reqCountCurrent > 0 {
				time.Sleep(time.Millisecond * 50)
			} else {
				stop <- true
				return
			}
		}
	}()
	go server.Serve(listener)
	<-stop
}
