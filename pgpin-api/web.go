package main

import (
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"time"
)

// Helpers.

// webRead reads the JSON request body into the given dataRef. It
// returns an error if the read fails for any reason.
func webRead(req *http.Request, dataRef interface{}) error {
	err := json.NewDecoder(req.Body).Decode(dataRef)
	if err != nil {
		err := &pgpinError{
			Id:         "bad-request",
			Message:    "malformed JSON body",
			HttpStatus: 400,
		}
		return err
	}
	return nil
}

// webRespond writes an HTTP response to the given resp, either
// according to status and data if err is nil, or according to err
// if it's non-nil. It will attempt to coerce err into a pgpinError
// and respond with an appropriate error message, falling back to
// a generic 500 error if it can't. All web responses should go
// through this function.
func webRespond(resp http.ResponseWriter, status int, data interface{}, err error) {
	if err != nil {
		pgerr, ok := err.(*pgpinError)
		if ok {
			status = pgerr.HttpStatus
			data = pgerr
		} else {
			log("web.error", "err=%+v", err)
			status = 500
			data = &map[string]string{"id": "internal-error", "message": "internal server error"}
		}
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(status)
	resp.Write(b)
	resp.Write([]byte("\n"))
}

// Middleware.

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

// Db endpoints.

func webDbList(resp http.ResponseWriter, req *http.Request) {
	dbs, err := dataDbList()
	webRespond(resp, 200, dbs, err)
}

func webDbAdd(resp http.ResponseWriter, req *http.Request) {
	db := &db{}
	err := webRead(req, db)
	if err == nil {
		db, err = dataDbAdd(db.Name, db.Url)
	}
	webRespond(resp, 201, db, err)
}

func webDbGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := dataDbGet(c.URLParams["id"])
	webRespond(resp, 200, db, err)
}

func webDbRemove(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := dataDbRemove(c.URLParams["id"])
	webRespond(resp, 200, db, err)
}

// Pin endpoints.

func webPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := dataPinList()
	webRespond(resp, 200, pins, err)
}

func webPinCreate(resp http.ResponseWriter, req *http.Request) {
	pin := &pin{}
	err := webRead(req, pin)
	if err == nil {
		pin, err = dataPinCreate(pin.DbId, pin.Name, pin.Query)
	}
	webRespond(resp, 201, pin, err)
}

func webPinGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := dataPinGet(c.URLParams["id"])
	webRespond(resp, 200, pin, err)
}

func webPinDelete(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := dataPinDelete(c.URLParams["id"])
	webRespond(resp, 200, pin, err)
}

// Misc endpoints.

type status struct {
	Message string `json:"message"`
}

func webStatus(resp http.ResponseWriter, req *http.Request) {
	err := DataTest()
	ok := &status{Message: "ok"}
	webRespond(resp, 200, ok, err)
}

func webNotFound(resp http.ResponseWriter, req *http.Request) {
	err := &pgpinError{
		Id:         "not-found",
		Message:    "not found",
		HttpStatus: 404,
	}
	webRespond(resp, 0, nil, err)
}

// Server builder.

func webStart() {
	log("web.start")
	DataStart()
	goji.Use(webLogging)
	goji.Get("/dbs", webDbList)
	goji.Post("/dbs", webDbAdd)
	goji.Get("/dbs/:id", webDbGet)
	goji.Delete("/dbs/:id", webDbRemove)
	goji.Get("/pins", webPinList)
	goji.Post("/pins", webPinCreate)
	goji.Get("/pins/:id", webPinGet)
	goji.Delete("/pins/:id", webPinDelete)
	goji.Get("/status", webStatus)
	goji.NotFound(webNotFound)
	port := env.Int("PORT")
	log("web.serve", "port=%d", port)
	goji.Serve()
}
