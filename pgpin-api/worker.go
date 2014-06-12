package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func WorkerExtractPgerror(err error) (*string, error) {
	pgerr, ok := err.(pq.PGError)
	if ok {
		msg := pgerr.Get('M')
		return &msg, nil
	}
	if err.Error() == "driver: bad connection" {
		msg := "could not connect to database"
		return &msg, nil
	}
	return nil, err
}

// WorkerCoerceType returns a coerced version of the raw
// database value in, which we get from scanning into
// interface{}s. We expect queries from the following
// Postgres types to result in the following return values:
// [Postgres] -> [Go: in] -> [Go: WorkerCoerceType'd]
// text          []byte      string
// ???
func WorkerCoerceType(in interface{}) interface{} {
	switch in := in.(type) {
	case []byte:
		return string(in)
	default:
		return in
	}
}

// WorkerQuery queries the pn db at pinDbUrl and updates the
// passed pin according to the results/errors. System errors
// are returned.
func WorkerQuery(p *Pin, pinDbUrl string) error {
	log.Printf("worker.query.start pin_id=%s", p.Id)
	pinDbConn := fmt.Sprintf("%s?application_name=pgpin.pin.%s", pinDbUrl, p.Id)
	pinDb, err := sql.Open("postgres", pinDbConn)
	if err != nil {
		p.ResultsError, err = WorkerExtractPgerror(err)
		return err
	}
	resultsRows, err := pinDb.Query(p.Query)
	if err != nil {
		p.ResultsError, err = WorkerExtractPgerror(err)
		return err
	}
	defer resultsRows.Close()
	resultsFieldsData, err := resultsRows.Columns()
	if err != nil {
		p.ResultsError, err = WorkerExtractPgerror(err)
		return err
	}
	resultsRowsData := make([][]interface{}, 0)
	resultsRowsSeen := 0
	for resultsRows.Next() {
		resultsRowsSeen += 1
		if resultsRowsSeen > DataPinResultsRowsMax {
			message := "too many rows in query results"
			p.ResultsError = &message
			return nil
		}
		resultsRowData := make([]interface{}, len(resultsFieldsData))
		resultsRowPointers := make([]interface{}, len(resultsFieldsData))
		for i, _ := range resultsRowData {
			resultsRowPointers[i] = &resultsRowData[i]
		}
		err := resultsRows.Scan(resultsRowPointers...)
		if err != nil {
			p.ResultsError, err = WorkerExtractPgerror(err)
			return err
		}
		for i, _ := range resultsRowData {
			resultsRowData[i] = WorkerCoerceType(resultsRowData[i])
		}
		resultsRowsData = append(resultsRowsData, resultsRowData)
	}
	err = resultsRows.Err()
	if err != nil {
		p.ResultsError, err = WorkerExtractPgerror(err)
		return err
	}
	p.ResultsFields = MustNewPgJson(resultsFieldsData)
	p.ResultsRows = MustNewPgJson(resultsRowsData)
	log.Printf("worker.query.finish pin_id=%s", p.Id)
	return nil
}

// WorkerProcess performs a processes an update on the given
// pin, running its query against its db and updating the
// system database accordingly. User-caused errors are
// reflected in the updated pin record and will not cause a
// returned error. System-caused errors are returned.
func WorkerProcess(p *Pin) error {
	log.Printf("worker.process.start pin_id=%s", p.Id)
	pinDbUrl, err := DataPinDbUrl(p)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	p.QueryStartedAt = &startedAt
	err = WorkerQuery(p, pinDbUrl)
	if err != nil {
		return err
	}
	finishedAt := time.Now()
	p.QueryFinishedAt = &finishedAt
	err = DataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.process.finish pin_id=%s", p.Id)
	return nil
}

// WorkerTick processes 1 pending pin, if such a pin is
// available. It returns true iff a pin is processed.
func WorkerTick() (bool, error) {
	p, err := DataPinReserve()
	if err != nil {
		return false, err
	}
	if p != nil {
		log.Printf("worker.tick.found pin_id=%s", p.Id)
		err = WorkerProcess(p)
		if err != nil {
			return false, err
		}
		err = DataPinRelease(p)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// WorkerTrap returns a chanel that will be populated when
// an INT or TERM signals is received.
func WorkerTrap() chan bool {
	sig := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Printf("worker.trap")
		done <- true
	}()
	return done
}

var WorkerCooloff = time.Millisecond*500

func WorkerCheckPanic() {
	err := recover()
	if err != nil {
		log.Printf("worker.panic: %s", err)
		log.Print(string(debug.Stack()))
		time.Sleep(WorkerCooloff)
	}
}

func WorkerHandleError(err error) {
	log.Printf("worker.error %s", err.Error())
	time.Sleep(WorkerCooloff)
}

func WorkerCheckExit(done chan bool) {
	select {
		case <-done:
			log.Printf("worker.exit")
			os.Exit(0)
		default:
	}
}

func WorkerLoop(done chan bool) {
	log.Printf("worker.loop")
	defer WorkerCheckPanic()
	processed, err := WorkerTick()
	if err != nil {
		WorkerHandleError(err)
	}
	WorkerCheckExit(done)
	if err == nil && !processed {
		time.Sleep(WorkerCooloff)
	}
}

func WorkerStart() {
	log.Printf("worker.start")
	done := WorkerTrap()
	for {
		WorkerLoop(done)
	}
}
