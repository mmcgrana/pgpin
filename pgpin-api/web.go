package main

import (
	"encoding/json"
	"errors"
	"github.com/zenazn/goji/bind"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// Constants.

var webTimeoutDuration = time.Second * 5

// Helpers.

// webRead reads the JSON request body into the given dataRef. It
// returns an error if the read fails for any reason.
func webRead(req *http.Request, dataRef interface{}) error {
	err := json.NewDecoder(req.Body).Decode(dataRef)
	if err != nil {
		err := &PgpinError{
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
// if it's non-nil. It will attempt to coerce err into a PgpinError
// and respond with an appropriate error message, falling back to
// a generic 500 error if it can't. All web responses should go
// through this function.
func webRespond(resp http.ResponseWriter, status int, data interface{}, err error) {
	if err != nil {
		pgerr, ok := err.(*PgpinError)
		if ok {
			status = pgerr.HttpStatus
			data = pgerr
		} else {
			log.Printf("web.error %+s", err.Error())
			status = 500
			data = &map[string]string{
				"id": "internal-error",
				"message": "internal server error",
			}
		}
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	resp.WriteHeader(status)
	_, err = resp.Write(b)
	if err != nil {
		log.Printf("web.ioerror %s", err.Error())
	}
	_, err = resp.Write([]byte("\n"))
	if err != nil {
		log.Printf("web.ioerror %s", err.Error())
	}
}

// Middleware.

func webJsoner(inner http.Handler) http.Handler {
	outer := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		inner.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(outer)
}

func webRecoverer(h http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("web.panic: %s", err)
				log.Print(string(debug.Stack()))
				webRespond(resp, 0, nil, &PgpinError{
					Id:         "internal-error",
					Message:    "internal server error",
					HttpStatus: 500,
				})
			}
		}()
		h.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(fn)
}

func webLogger(inner http.Handler) http.Handler {
	outer := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		path := r.URL.Path
		log.Printf("web.request.start method=%s path=%s", method, path)
		inner.ServeHTTP(w, r)
		elapsed := float64(time.Since(start)) / 1000000.0
		log.Printf("web.request.finish method=%s path=%s elapsed=%f", method, path, elapsed)
	}
	return http.HandlerFunc(outer)
}

func webTimer(timeout time.Duration) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		data := &map[string]string{
			"id":      "request-timeout",
			"message": "request timed out",
		}
		body, err := json.MarshalIndent(data, "", "  ")
		must(err)
		return http.TimeoutHandler(inner, timeout, string(body)+"\n")
	}
}

// Db endpoints.

func webDbList(resp http.ResponseWriter, req *http.Request) {
	dbs, err := dataDbList()
	webRespond(resp, 200, dbs, err)
}

func webDbAdd(resp http.ResponseWriter, req *http.Request) {
	db := &Db{}
	err := webRead(req, db)
	if err == nil {
		db, err = dataDbAdd(db.Name, db.Url)
	}
	webRespond(resp, 201, db, err)
}

func webDbUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	dbUpdate := &Db{}
	db := &Db{}
	err := webRead(req, dbUpdate)
	if err == nil {
		db, err = dataDbGet(c.URLParams["id"])
		if err == nil {
			db.Name = dbUpdate.Name
			err = dataDbUpdate(db)
		}
	}
	webRespond(resp, 200, db, err)
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
	pin := &Pin{}
	err := webRead(req, pin)
	if err == nil {
		pin, err = dataPinCreate(pin.DbId, pin.Name, pin.Query)
	}
	webRespond(resp, 201, pin, err)
}

func webPinUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	pinUpdate := &Pin{}
	pin := &Pin{}
	err := webRead(req, pinUpdate)
	if err == nil {
		pin, err = dataPinGet(c.URLParams["id"])
		if err == nil {
			pin.Name = pinUpdate.Name
			err = dataPinUpdate(pin)
		}
	}
	webRespond(resp, 200, pin, err)
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

type Status struct {
	Message string `json:"message"`
}

func webStatus(resp http.ResponseWriter, req *http.Request) {
	err := dataConn.Ping()
	status := &Status{Message: "ok"}
	webRespond(resp, 200, status, err)
}

func webError(resp http.ResponseWriter, req *http.Request) {
	err := errors.New("a problem occurred")
	webRespond(resp, 0, nil, err)
}

func webPanic(resp http.ResponseWriter, req *http.Request) {
	panic("panic")
}

func webTimeout(resp http.ResponseWriter, req *http.Request) {
	time.Sleep(webTimeoutDuration + time.Second)
	status := &Status{Message: "late"}
	webRespond(resp, 200, status, nil)
}

func webNotFound(resp http.ResponseWriter, req *http.Request) {
	err := &PgpinError{
		Id:         "not-found",
		Message:    "not found",
		HttpStatus: 404,
	}
	webRespond(resp, 0, nil, err)
}

// Server builder.

var webMux *web.Mux

func webBuild() {
	webMux = web.New()
	webMux.Use(webJsoner)
	webMux.Use(webLogger)
	webMux.Use(webTimer(webTimeoutDuration))
	webMux.Use(webRecoverer)
	webMux.Get("/dbs", webDbList)
	webMux.Post("/dbs", webDbAdd)
	webMux.Put("/dbs/:id", webDbUpdate)
	webMux.Get("/dbs/:id", webDbGet)
	webMux.Delete("/dbs/:id", webDbRemove)
	webMux.Get("/pins", webPinList)
	webMux.Post("/pins", webPinCreate)
	webMux.Put("/pins/:id", webPinUpdate)
	webMux.Get("/pins/:id", webPinGet)
	webMux.Delete("/pins/:id", webPinDelete)
	webMux.Get("/status", webStatus)
	webMux.Get("/error", webError)
	webMux.Get("/panic", webPanic)
	webMux.Get("/timeout", webTimeout)
	webMux.NotFound(webNotFound)
}

func webStart() {
	log.Print("web.start")
	webBuild()
	listener := bind.Default()
	bind.Ready()
	must(graceful.Serve(listener, webMux))
	graceful.Wait()
}
