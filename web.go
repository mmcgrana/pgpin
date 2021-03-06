package main

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/graceful"
	"github.com/zenazn/goji/web"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

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
				"id":      "internal-error",
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

func WebRequestIder(c *web.C, h http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		var requestId string
		given := req.Header.Get("X-Request-Id")
		if given != "" {
			requestId = given
		}
		given = req.Header.Get("Request-Id")
		if given != "" {
			requestId = given
		}
		if requestId == "" {
			requestId = uuid.New()
		}
		resp.Header().Set("Request-Id", requestId)
		h.ServeHTTP(resp, req)
	}
	return http.HandlerFunc(fn)
}

func WebLogger(c *web.C, inner http.Handler) http.Handler {
	outer := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		path := r.URL.Path
		requestId := w.Header().Get("Request-Id")
		log.Printf("web.request.start request_id=%s method=%s path=%s", requestId, method, path)
		inner.ServeHTTP(w, r)
		elapsed := float64(time.Since(start)) / 1000000.0
		log.Printf("web.request.finish request_id=%s method=%s path=%s elapsed=%f", requestId, method, path, elapsed)
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

// Db endpoints.

type DbSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func WebDbList(resp http.ResponseWriter, req *http.Request) {
	dbs, err := DbList("")
	dbSlims := []*DbSlim{}
	for _, db := range dbs {
		dbSlims = append(dbSlims, &DbSlim{Id: db.Id, Name: db.Name})
	}
	WebRespond(resp, 200, dbSlims, err)
}

func WebDbCreate(resp http.ResponseWriter, req *http.Request) {
	db := &Db{}
	err := WebRead(req, db)
	if err == nil {
		db, err = DbCreate(db.Name, db.Url)
	}
	WebRespond(resp, 201, db, err)
}

func WebDbUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	dbUpdate := &Db{}
	db := &Db{}
	err := WebRead(req, dbUpdate)
	if err == nil {
		db, err = DbGet(c.URLParams["id"])
		if err == nil {
			if dbUpdate.Name != "" {
				db.Name = dbUpdate.Name
			}
			if dbUpdate.Url != "" {
				db.Url = dbUpdate.Url
			}
			err = DbUpdate(db)
		}
	}
	WebRespond(resp, 200, db, err)
}

func WebDbGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := DbGet(c.URLParams["id"])
	WebRespond(resp, 200, db, err)
}

func WebDbDelete(c web.C, resp http.ResponseWriter, req *http.Request) {
	db, err := DbDelete(c.URLParams["id"])
	WebRespond(resp, 200, db, err)
}

// Pin endpoints.

type PinSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func WebPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := PinList("")
	pinSlims := []*PinSlim{}
	for _, pin := range pins {
		pinSlims = append(pinSlims, &PinSlim{Id: pin.Id, Name: pin.Name})
	}
	WebRespond(resp, 200, pinSlims, err)
}

func WebPinCreate(resp http.ResponseWriter, req *http.Request) {
	pin := &Pin{}
	err := WebRead(req, pin)
	if err == nil {
		pin, err = PinCreate(pin.DbId, pin.Name, pin.Query)
	}
	WebRespond(resp, 201, pin, err)
}

func WebPinUpdate(c web.C, resp http.ResponseWriter, req *http.Request) {
	pinUpdate := &Pin{}
	pin := &Pin{}
	err := WebRead(req, pinUpdate)
	if err == nil {
		pin, err = PinGet(c.URLParams["id"])
		if err == nil {
			if pinUpdate.Name != "" {
				pin.Name = pinUpdate.Name
			}
			if pinUpdate.Query != "" {
				pin.Query = pinUpdate.Query
			}
			err = PinUpdate(pin)
		}
	}
	WebRespond(resp, 200, pin, err)
}

func WebPinGet(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := PinGet(c.URLParams["id"])
	WebRespond(resp, 200, pin, err)
}

func WebPinDelete(c web.C, resp http.ResponseWriter, req *http.Request) {
	pin, err := PinDelete(c.URLParams["id"])
	WebRespond(resp, 200, pin, err)
}

// Misc endpoints.

type Status struct {
	Message string `json:"message"`
}

func WebStatus(resp http.ResponseWriter, req *http.Request) {
	err := PgConn.Ping()
	status := &Status{Message: "ok"}
	WebRespond(resp, 200, status, err)
}

func WebTriggerError(resp http.ResponseWriter, req *http.Request) {
	err := errors.New("a problem occurred")
	WebRespond(resp, 0, nil, err)
}

func WebTriggerPanic(resp http.ResponseWriter, req *http.Request) {
	panic("panic")
}

func WebTriggerSleep(resp http.ResponseWriter, req *http.Request) {
	time.Sleep(ConfigWebTimeout - (100 * time.Millisecond))
	status := &Status{Message: "sleepy"}
	WebRespond(resp, 200, status, nil)
}

func WebTriggerTimeout(resp http.ResponseWriter, req *http.Request) {
	time.Sleep(ConfigWebTimeout + (100 * time.Millisecond))
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
	WebMux.Use(WebRequestIder)
	WebMux.Use(WebLogger)
	WebMux.Use(WebTimer(ConfigWebTimeout))
	WebMux.Use(WebRecoverer)
	WebMux.Get("/v1/dbs", WebDbList)
	WebMux.Post("/v1/dbs", WebDbCreate)
	WebMux.Put("/v1/dbs/:id", WebDbUpdate)
	WebMux.Get("/v1/dbs/:id", WebDbGet)
	WebMux.Delete("/v1/dbs/:id", WebDbDelete)
	WebMux.Get("/v1/pins", WebPinList)
	WebMux.Post("/v1/pins", WebPinCreate)
	WebMux.Put("/v1/pins/:id", WebPinUpdate)
	WebMux.Get("/v1/pins/:id", WebPinGet)
	WebMux.Delete("/v1/pins/:id", WebPinDelete)
	WebMux.Get("/status", WebStatus)
	WebMux.Get("/error", WebTriggerError)
	WebMux.Get("/panic", WebTriggerPanic)
	WebMux.Get("/sleep", WebTriggerSleep)
	WebMux.Get("/timeout", WebTriggerTimeout)
	WebMux.NotFound(WebNotFound)
}

func WebStart() {
	log.Print("web.start")
	PgStart()
	RedisStart()
	WebBuild()
	addr := fmt.Sprintf(":%d", ConfigWebPort)
	graceful.Run(addr, ConfigWebDrainInterval, WebMux)
}
