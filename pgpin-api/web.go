package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"github.com/darkhelmet/env"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func webPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := dataPinList()
	if err != nil {
		webErr(resp, err)
		return
	}
	httpWriteJson(resp, 200, pins)
}

func webPinCreate(resp http.ResponseWriter, req *http.Request) {
	pinReq := pin{}
	err := httpReadJson(req, &pinReq)
	if err != nil {
		err = pgpinError{Id: "bad-request", Message: "malformed JSON body"}
		webErr(resp, err)
		return
	}
	pin, err := dataPinCreate(pinReq.DbId, pinReq.Name, pinReq.Query)
	if err != nil {
		webErr(resp, err)
		return
	}
	httpWriteJson(resp, 200, pin)
}

func webPinGet(resp http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	pin, err := dataPinGet(id)
	if err != nil {
		webErr(resp, err)
		return
	}
	httpWriteJson(resp, 200, pin)
}

func webPinDestroy(resp http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
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
	httpWriteJson(resp, 200, pin)
}

func webStatus(resp http.ResponseWriter, req *http.Request) {
	err := dataTest()
	if err != nil {
		webErr(resp, err)
		return
	}
	httpWriteJson(resp, 200, &map[string]string{"message": "ok"})
}

func webNotFound(resp http.ResponseWriter, req *http.Request) {
	err := pgpinError{Id: "not-found", Message: "not found"}
	webErr(resp, err)
}

func webErr(resp http.ResponseWriter, err error) {
	fmt.Println("error:", err)
	switch err.(type) {
	case pgpinError:
		httpWriteJson(resp, 500, err)
	default:
		httpWriteJson(resp, 500, &map[string]string{"id": "internal-error", "message": "internal server error"})
	}
}

func webWrapAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {	
		f(res, req)
	}
}

func webRouterHandler() http.HandlerFunc {
	router := mux.NewRouter()
	router.HandleFunc("/pins", webPinList).Methods("GET")
	router.HandleFunc("/pins", webPinCreate).Methods("POST")
	router.HandleFunc("/pins/{id}", webPinGet).Methods("GET")
	router.HandleFunc("/pins/{id}", webPinDestroy).Methods("DELETE")
	router.HandleFunc("/status", webStatus).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(webNotFound)
	return func(res http.ResponseWriter, req *http.Request) {
		router.ServeHTTP(res, req)
	}
}

func webTrap() {
	log("web.trap.set")
	trap := make(chan os.Signal)
	go func() {
		<- trap
		log("web.exit")
		os.Exit(0)
	}()
	signal.Notify(trap, syscall.SIGINT, syscall.SIGTERM)
}

type webStatusingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *webStatusingResponseWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

func webWrapLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		method := req.Method
		path := req.URL.Path
		log("web.request.start method=%s path=%s", method, path)
		wres := webStatusingResponseWriter{-1, res}
		f(&wres, req)
		elapsed := float64(time.Since(start)) / 1000000.0
		log("web.request.finish method=%s path=%s status=%d elapsed=%f", method, path, wres.status, elapsed)
	}
}

func web() {
	dataInit()
	log("web.start")
	handler := webRouterHandler()
	handler = webWrapAuth(handler)
	handler = webWrapLogging(handler)
	webTrap()
	port := env.Int("PORT")
	log("web.serve port=%d", port)	
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	if err != nil {
		panic(err)
	}
}
