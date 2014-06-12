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

var WebTimeoutDuration = time.Second * 5

// Helpers.

// WebRead reads the JSON request body into the given dataRef. It
// returns an error if the read fails for any reason.
func WebRead(req *http.Request, dataRef interface{}) error {
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

// WebRespond writes an HTTP response to the given resp, either
// according to status and data if err is nil, or according to err
// if it's non-nil. It will attempt to coerce err into a PgpinError
// and respond with an appropriate error message, falling back to
// a generic 500 error if it can't. All web responses should go
// through this function.
func WebRespond(resp http.ResponseWriter, status int, data interface{}, err error) {
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

func WebJsoner(inner http.Handler) http.Handler {
	outer := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		inner.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(outer)
}

func WebRecoverer(h http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("web.panic: %s", err)
				log.Print(string(debug.Stack()))
				WebRespond(resp, 0, nil, &PgpinError{
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

func WebLogger(inner http.Handler) http.Handler {
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

func WebTimer(timeout time.Duration) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		data := &map[string]string{
			"id":      "request-timeout",
			"message": "request timed out",
		}
		body, err := json.MarshalIndent(data, "", "  ")
		Must(err)
		return http.TimeoutHandler(inner, timeout, string(body)+"\n")
	}
}

// Db endpoints.

func WebDbList(resp http.ResponseWriter, req *http.Request) {
	dbs, err := DataDbList()
	WebRespond(resp, 200, dbs, err)
}

func WebDbAdd(resp http.ResponseWriter, req *http.Request) {
	db := &Db{}
	err := WebRead(req, db)
	if err == nil {
		db, err = DataDbAdd(db.Name, db.Url)
	}
	WebRespond(resp, 201, db, err)
}

func WebDbUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	dbUpdate := &Db{}
	db := &Db{}
	err := WebRead(req, dbUpdate)
	if err == nil {
		db, err = DataDbGet(c.URLParams["id"])
		if err == nil {
			if dbUpdate.Name != "" {
				db.Name = dbUpdate.Name
			}
			if dbUpdate.Url != "" {
				db.Url = dbUpdate.Url
			}
			err = DataDbUpdate(db)
		}
	}
	WebRespond(resp, 200, db, err)
}

func WebDbGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := DataDbGet(c.URLParams["id"])
	WebRespond(resp, 200, db, err)
}

func WebDbRemove(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := DataDbRemove(c.URLParams["id"])
	WebRespond(resp, 200, db, err)
}

// Pin endpoints.

func WebPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := DataPinList()
	WebRespond(resp, 200, pins, err)
}

func WebPinCreate(resp http.ResponseWriter, req *http.Request) {
	pin := &Pin{}
	err := WebRead(req, pin)
	if err == nil {
		pin, err = DataPinCreate(pin.DbId, pin.Name, pin.Query)
	}
	WebRespond(resp, 201, pin, err)
}

func WebPinUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	pinUpdate := &Pin{}
	pin := &Pin{}
	err := WebRead(req, pinUpdate)
	if err == nil {
		pin, err = DataPinGet(c.URLParams["id"])
		if err == nil {
			if pinUpdate.Name != "" {
				pin.Name = pinUpdate.Name
			}
			if pinUpdate.Query != "" {
				pin.Query = pinUpdate.Query
			}
			err = DataPinUpdate(pin)
		}
	}
	WebRespond(resp, 200, pin, err)
}

func WebPinGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := DataPinGet(c.URLParams["id"])
	WebRespond(resp, 200, pin, err)
}

func WebPinDelete(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := DataPinDelete(c.URLParams["id"])
	WebRespond(resp, 200, pin, err)
}

// Misc endpoints.

type Status struct {
	Message string `json:"message"`
}

func WebStatus(resp http.ResponseWriter, req *http.Request) {
	err := DataConn.Ping()
	status := &Status{Message: "ok"}
	WebRespond(resp, 200, status, err)
}

func WebError(resp http.ResponseWriter, req *http.Request) {
	err := errors.New("a problem occurred")
	WebRespond(resp, 0, nil, err)
}

func WebPanic(resp http.ResponseWriter, req *http.Request) {
	panic("panic")
}

func WebTimeout(resp http.ResponseWriter, req *http.Request) {
	time.Sleep(WebTimeoutDuration + time.Second)
	status := &Status{Message: "late"}
	WebRespond(resp, 200, status, nil)
}

func WebNotFound(resp http.ResponseWriter, req *http.Request) {
	err := &PgpinError{
		Id:         "not-found",
		Message:    "not found",
		HttpStatus: 404,
	}
	WebRespond(resp, 0, nil, err)
}

// Server builder.

var WebMux *web.Mux

func WebBuild() {
	WebMux = web.New()
	WebMux.Use(WebJsoner)
	WebMux.Use(WebLogger)
	WebMux.Use(WebTimer(WebTimeoutDuration))
	WebMux.Use(WebRecoverer)
	WebMux.Get("/dbs", WebDbList)
	WebMux.Post("/dbs", WebDbAdd)
	WebMux.Put("/dbs/:id", WebDbUpdate)
	WebMux.Get("/dbs/:id", WebDbGet)
	WebMux.Delete("/dbs/:id", WebDbRemove)
	WebMux.Get("/pins", WebPinList)
	WebMux.Post("/pins", WebPinCreate)
	WebMux.Put("/pins/:id", WebPinUpdate)
	WebMux.Get("/pins/:id", WebPinGet)
	WebMux.Delete("/pins/:id", WebPinDelete)
	WebMux.Get("/status", WebStatus)
	WebMux.Get("/error", WebError)
	WebMux.Get("/panic", WebPanic)
	WebMux.Get("/timeout", WebTimeout)
	WebMux.NotFound(WebNotFound)
}

func WebStart() {
	log.Print("web.start")
	WebBuild()
	listener := bind.Default()
	bind.Ready()
	Must(graceful.Serve(listener, WebMux))
	graceful.Wait()
}
