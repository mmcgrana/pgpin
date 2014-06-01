package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getToken(resp http.ResponseWriter, req *http.Request) (string, bool) {
	_, token, ok := getAuth(req)
	if !ok {
		err := notAuthorizedError{Message: "authorization required"}
		writeErr(resp, err)
		return "", false
	}
	return token, true
}

func getResources(resp http.ResponseWriter, req *http.Request) {
	token, ok := getToken(resp, req)
	if !ok {
		return
	}
	resources, err := dataGetResources(token)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, resources)
}

func getPins(resp http.ResponseWriter, req *http.Request) {
	token, ok := getToken(resp, req)
	if !ok {
		return
	}
	pins, err := dataGetPins(token)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pins)
}

func createPin(resp http.ResponseWriter, req *http.Request) {
	token, ok := getToken(resp, req)
	if !ok {
		return
	}
	pinReq := pin{}
	ok = readJson(resp, req, &pinReq)
	if !ok {
		err := malformedError{Message: "invalid body"}
		writeErr(resp, err)
		return
	}
	pin, err := dataCreatePin(token, pinReq.ResourceId, pinReq.Name, pinReq.Sql)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func getPin(resp http.ResponseWriter, req *http.Request) {
	token, ok := getToken(resp, req)
	if !ok {
		return
	}
	id := param(req, "id")
	pin, err := dataGetPin(token, id)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func deletePin(resp http.ResponseWriter, req *http.Request) {
	token, ok := getToken(resp, req)
	if !ok {
		return
	}
	id := param(req, "id")
	pin, err := dataGetPin(token, id)
	if err != nil {
		writeErr(resp, err)
		return
	}
	err = dataDeletePin(pin)
	if err != nil {
		writeErr(resp, err)
		return
	}
	writeJson(resp, 200, pin)
}

func getStatus(resp http.ResponseWriter, req *http.Request) {
	err := dataTest()
	if err != nil {
		writeErr(resp, err)
		return
	}
	time.Sleep(time.Second * 5)
	writeJson(resp, 200, &stringMap{"message": "ok"})
}

func notFound(resp http.ResponseWriter, req *http.Request) {
	err := notFoundError{Message: "not found"}
	writeErr(resp, err)
}

func writeErr(resp http.ResponseWriter, err error) {
	fmt.Println("error:", err)
	switch e := err.(type) {
	case malformedError:
		writeJson(resp, 400, &stringMap{"message": e.Error()})
	case notAuthorizedError:
		writeJson(resp, 401, &stringMap{"message": e.Error()})
	case invalidError:
		writeJson(resp, 403, &stringMap{"message": e.Error()})
	case notFoundError:
		writeJson(resp, 404, &stringMap{"message": e.Error()})
	default:
		writeJson(resp, 500, &stringMap{"message": "internal server error"})
	}
}

func router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/v1/resources", getResources).Methods("GET")
	router.HandleFunc("/v1/pins", getPins).Methods("GET")
	router.HandleFunc("/v1/pins", createPin).Methods("POST")
	router.HandleFunc("/v1/pins/{id}", getPin).Methods("GET")
	router.HandleFunc("/v1/pins/{id}", deletePin).Methods("DELETE")
	router.HandleFunc("/v1/status", getStatus).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(notFound)
	return router
}

func webTrap() chan os.Signal {
	traps := make(chan os.Signal)
	sigs := make(chan os.Signal)
	go func() {
		s := <- traps
		log("key=web.trap")
		sigs <- s
	}()
	signal.Notify(traps, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

func web() {
	dataInit()
	log("key=web.start")
	handler := routerHandlerFunc(router())
	handler = wrapLogging(handler)
	sigs := webTrap()
	port := confPort()
	log("key=web.serve port=%d", port)
	httpServeGraceful(handler, port, sigs)
	log("key=web.exit")
}
