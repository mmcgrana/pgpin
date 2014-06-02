package main

import (
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"time"
)

func webRespond(resp http.ResponseWriter, status int, data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(status)
	resp.Write(b)
	resp.Write([]byte("\n"))
}

func webPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := dataPinList()
	if err != nil {
		webErr(resp, err)
		return
	}
	webRespond(resp, 200, pins)
}

func webPinCreate(resp http.ResponseWriter, req *http.Request) {
	pinReq := pin{}
	err := json.NewDecoder(req.Body).Decode(&pinReq)
	if err != nil {
		err = &pgpinError{
			Id:         "bad-request",
			Message:    "malformed JSON body",
			HttpStatus: 400,
		}
		webErr(resp, err)
		return
	}
	pin, err := dataPinCreate(pinReq.DbId, pinReq.Name, pinReq.Query)
	if err != nil {
		webErr(resp, err)
		return
	}
	webRespond(resp, 200, pin)
}

func webPinGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	id := c.URLParams["id"]
	pin, err := dataPinGet(id)
	if err != nil {
		webErr(resp, err)
		return
	}
	webRespond(resp, 200, pin)
}

func webPinDestroy(c web.C, resp http.ResponseWriter, req *http.Request) {
	id := c.URLParams["id"]
	pin, err := dataPinGet(id)
	if err != nil {
		webErr(resp, err)
		return
	}
	err = dataPinDelete(pin)
	if err != nil {
		webErr(resp, err)
		return
	}
	webRespond(resp, 200, pin)
}

func webStatus(resp http.ResponseWriter, req *http.Request) {
	err := dataTest()
	if err != nil {
		webErr(resp, err)
		return
	}
	webRespond(resp, 200, &map[string]string{"message": "ok"})
}

func webNotFound(resp http.ResponseWriter, req *http.Request) {
	err := &pgpinError{
		Id:         "not-found",
		Message:    "not found",
		HttpStatus: 404,
	}
	webErr(resp, err)
}

func webErr(resp http.ResponseWriter, err error) {
	pgpinerr, ok := err.(*pgpinError)
	if ok {
		webRespond(resp, pgpinerr.HttpStatus, pgpinerr)
	} else {
		log("web.error", "err=%+v", err)
		webRespond(resp, 500, &map[string]string{"id": "internal-error", "message": "internal server error"})
	}
}

func webLogging(inner http.Handler) http.Handler {
	outer := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		path := r.URL.Path
		log("web.request.start", "method=%s path=%s", method, path)
		inner.ServeHTTP(w, r)
		elapsed := float64(time.Since(start)) / 1000000.0
		log("web.request.finish", "method=%s path=%s elapsed=%f", method, path, elapsed)
	}
	return http.HandlerFunc(outer)
}

func webStart() {
	log("web.start")
	dataStart()
	goji.Get("/pins", webPinList)
	goji.Post("/pins", webPinCreate)
	goji.Get("/pins/:id", webPinGet)
	goji.Delete("/pins/:id", webPinDestroy)
	goji.Get("/status", webStatus)
	goji.NotFound(webNotFound)
	goji.Use(webLogging)
	port := env.Int("PORT")
	log("web.serve", "port=%d", port)
	goji.Serve()
}
