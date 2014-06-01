package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"github.com/darkhelmet/env"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func webPinList(resp http.ResponseWriter, req *http.Request) {
	pins, err := dataPinList()
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pins)
}

func webPinCreate(resp http.ResponseWriter, req *http.Request) {
	pinReq := pin{}
	err := readJson(req, &pinReq)
	if err != nil {
		err = pgpinError{Id: "bad-request", Message: "malformed JSON body"}
		writeErr(resp, err)
		return
	}
	pin, err := dataPinCreate(pinReq.DbId, pinReq.Name, pinReq.Query)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func webPinGet(resp http.ResponseWriter, req *http.Request) {
	id := param(req, "id")
	pin, err := dataPinGet(id)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func webPinDestroy(resp http.ResponseWriter, req *http.Request) {
	id := param(req, "id")
	pin, err := dataPinGet(id)
	if err != nil {
		writeErr(resp, err)
		return
	}
	err = dataPinDelete(pin)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func webApiStatus(resp http.ResponseWriter, req *http.Request) {
	err := dataTest()
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, &map[string]string{"message": "ok"})
}

func notFound(resp http.ResponseWriter, req *http.Request) {
	err := pgpinError{Id: "not-found", Message: "not found"}
	writeErr(resp, err)
}

func writeErr(resp http.ResponseWriter, err error) {
	fmt.Println("error:", err)
	switch err.(type) {
	case pgpinError:
		writeJson(resp, 500, err)
	default:
		writeJson(resp, 500, &map[string]string{"id": "internal-error", "message": "internal server error"})
	}
}

func wrapAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		
	}
}

func webRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/pins", webPinList).Methods("GET")
	router.HandleFunc("/pins", webPinCreate).Methods("POST")
	router.HandleFunc("/pins/{id}", webPinGet).Methods("GET")
	router.HandleFunc("/pins/{id}", webPinDestroy).Methods("DELETE")
	router.HandleFunc("/api-status", webApiStatus).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(notFound)
	return router
}

func webTrap() chan os.Signal {
	traps := make(chan os.Signal)
	sigs := make(chan os.Signal)
	go func() {
		s := <- traps
		log("web.trap")
		sigs <- s
	}()
	signal.Notify(traps, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

func web() {
	dataInit()
	log("web.start")
	handler := routerHandlerFunc(webRouter())
	handler = wrapAuth(handler)
	handler = wrapLogging(handler)
	sigs := webTrap()
	port := env.Int("PORT")
	log("web.serve port=%d", port)
	httpServeGraceful(handler, port, sigs)
	log("web.exit")
}
