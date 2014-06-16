package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/jrallison/go-workers"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

// Setup and teardown.

func init() {
	if ConfigTestLogs {
		log.SetOutput(ioutil.Discard)
	}
	ConfigDatabaseUrl = env.String("TEST_DATABASE_URL")
	ConfigRedisUrl = env.String("TEST_REDIS_URL")
	WebBuild()
	DataStart()
	QueueStart()
	clear()
}

func clear() {
	_, err := DataConn.Exec("DELETE from pins")
	Must(err)
	_, err = DataConn.Exec("DELETE from dbs")
	Must(err)
	conn := workers.Config.Pool.Get()
	_, err = conn.Do("flushdb")
	defer conn.Close()
	Must(err)
}

// Helpers.

func asReader(body string) io.Reader {
	return bytes.NewReader([]byte(body))
}

func mustRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	Must(err)
	res := httptest.NewRecorder()
	WebMux.ServeHTTP(res, req)
	return res
}

func mustDecode(res *httptest.ResponseRecorder, data interface{}) {
	Must(json.NewDecoder(res.Body).Decode(data))
}

func mustDataDbAdd(name string, url string) *Db {
	db, err := DataDbAdd(name, url)
	Must(err)
	return db
}

func mustDataPinCreate(dbId string, name string, query string) *Pin {
	pin, err := DataPinCreate(dbId, name, query)
	Must(err)
	return pin
}

func mustDataPinGet(id string) *Pin {
	pin, err := DataPinGet(id)
	Must(err)
	return pin
}

func mustWorkerTick() {
	var message *workers.Msg
	fetcher := workers.NewFetch("queue:pins", make(chan *workers.Msg))
	go fetcher.Fetch()
	defer fetcher.Close()
	select {
	case message = <-fetcher.Messages():
	case <-time.After(10 * time.Millisecond):
	}
	if message != nil {
		WorkerProcessWrapper(message)
		fetcher.Acknowledge(message)
	}
}

func mustSchedulerTick() {
	Must(SchedulerTick())
}

func mustCanonicalizeJson(in []byte) string {
	data := make([]interface{}, 0)
	Must(json.Unmarshal(in, &data))
	out, err := json.Marshal(data)
	Must(err)
	return string(out)

}
