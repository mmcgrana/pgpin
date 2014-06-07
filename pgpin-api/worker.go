package main

import (
	"database/sql"
	"github.com/lib/pq"
	"log"
	"time"
)

var resultsRowsMax = 10000

func workerExtractPgerror(err error) (*string, error) {
	pgerr, ok := err.(pq.PGError)
	if ok {
	        msg := pgerr.Get('M')
	        return &msg, nil
	}
	return nil, err
}

// workerCoerceType returns a coerced version of the raw
// database value in, which we get from scanning into
// interface{}s. We expect queries from the following
// Postgres types to result in the following return values:
// [Postgres] -> [Go: in] -> [Go: workerCoerceType'd]
// text          []byte      string
// ???
func workerCoerceType(in interface{}) interface{} {
	switch in := in.(type) {
	case []byte:
		return string(in)
	default:
		return in
	}
}

// workerQuery queries the pn db at pinDbUrl and updates the
// passed pin according to the results/errors. System errors
// are returned.
func workerQuery(p *pin, pinDbUrl string) error {
	log.Print("worker.query.start pin_id=%s", p.Id)
	pinDb, err := sql.Open("postgres", pinDbUrl)
	if err != nil {
		p.ResultsError, err = workerExtractPgerror(err)
		return err
	}
	resultsRows, err := pinDb.Query(p.Query)
	defer resultsRows.Close()
	if err != nil {
		p.ResultsError, err = workerExtractPgerror(err)
		return err
	}
	resultsFieldsData, err := resultsRows.Columns()
	if err != nil {
		p.ResultsError, err = workerExtractPgerror(err)
		return err
	}
	resultsRowsData := make([][]interface{}, 0)
	resultsRowsSeen := 0
	for resultsRows.Next() {
		resultsRowsSeen += 1
		if resultsRowsSeen > resultsRowsMax {
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
			p.ResultsError, err = workerExtractPgerror(err)
			return err
		}
		for i, _ := range resultsRowData {
			resultsRowData[i] = workerCoerceType(resultsRowData[i])
		}
		resultsRowsData = append(resultsRowsData, resultsRowData)
	}
	err = resultsRows.Err()
	if err != nil {
		p.ResultsError, err = workerExtractPgerror(err)
		return err
	}
	p.ResultsFields = MustNewPgJson(resultsFieldsData)
	p.ResultsRows = MustNewPgJson(resultsRowsData)
	log.Print("worker.query.finish pin_id=%s", p.Id)
	return nil
}

// workerProcess performs a processes an update on the given
// pin, running its query against its db and updating the
// system database accordingly. User-caused errors are
// reflected in the updated pin record and will not cause a
// returned error. System-caused errors are returned.
func workerProcess(p *pin) error {
	log.Printf("worker.process.start pin_id=%s", p.Id)
	pinDbUrl, err := dataPinDbUrl(p)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	p.QueryStartedAt = &startedAt
	err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	err = workerQuery(p, pinDbUrl)
	if err != nil {
		return err
	}
	finishedAt := time.Now()
	p.QueryFinishedAt = &finishedAt
	err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.process.finish pin_id=%s", p.Id)
	return nil
}

// workerTick processes 1 pending pin, if such a pin is
// available. It returns true iff a pin is processed.
func workerTick() (bool, error) {
	p, err := dataPinForQuery()
	if err != nil {
		return false, err
	}
	if p != nil {
		log.Print("worker.tick.found pin_id=%s", p.Id)
		err = workerProcess(p)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func workerStart() {
	log.Print("worker.start")
	for {
		processed, err := workerTick()
		if err != nil {
			log.Print("worker.error %s", err.Error())
			panic(err)
		}
		if !processed {
			time.Sleep(time.Millisecond)
		}
	}
}
