package main

import (
	"database/sql"
	"fmt"
	"github.com/jrallison/go-workers"
	"github.com/lib/pq"
	"log"
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

// WorkerQuery queries the pin db at pinDbUrl and updates the
// passed pin according to the results/errors. System errors
// are returned.
func WorkerQuery(p *Pin, pinDbUrl string) error {
	log.Printf("worker.query.start pin_id=%s", p.Id)
	applicationName := fmt.Sprintf("pgpin.pin.%s", p.Id)
	pinDbConn := fmt.Sprintf("%s?application_name=%s&statement_timeout=%d&connect_timeout=%d",
		pinDbUrl, applicationName, ConfigPinStatementTimeout/time.Millisecond, ConfigDatabaseConnectTimeout/time.Millisecond)
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
	defer func() { Must(resultsRows.Close()) }()
	resultsFieldsData, err := resultsRows.Columns()
	if err != nil {
		p.ResultsError, err = WorkerExtractPgerror(err)
		return err
	}
	resultsRowsData := make([][]interface{}, 0)
	resultsRowsSeen := 0
	for resultsRows.Next() {
		resultsRowsSeen += 1
		if resultsRowsSeen > ConfigPinResultsRowsMax {
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

func WorkerProcess(jobId string, pinId string) error {
	log.Printf("worker.job.start job_id=%s pin_id=%s", jobId, pinId)
	pin, err := DataPinGet(pinId)
	if err != nil {
		return err
	}
	pinDbUrl, err := DataPinDbUrl(pin)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	pin.QueryStartedAt = &startedAt
	err = WorkerQuery(pin, pinDbUrl)
	if err != nil {
		return err
	}
	finishedAt := time.Now()
	pin.QueryFinishedAt = &finishedAt
	err = DataPinUpdate(pin)
	if err != nil {
		return err
	}
	log.Printf("worker.job.finish job_id=%s pin_id=%s", jobId, pinId)
	return nil
}

func WorkerProcessWrapper(msg *workers.Msg) {
	jobId := msg.Jid()
	pinId, err := msg.Args().String()
	Must(err)
	err = WorkerProcess(jobId, pinId)
	if err != nil {
		log.Printf("worker.job.error job_id=%s pin_id=%s %s", jobId, pinId, err)
	}
}

func WorkerStart() {
	log.Printf("worker.start")
	PgStart()
	RedisStart()
	workers.Process("pins", WorkerProcessWrapper, 5)
	workers.Run()
}
