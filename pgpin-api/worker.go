package main

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

var resultsRowsMax = 10000

func workerPoll() (*pin, error) {
	pin, err := dataPinForQuery()
	if err != nil {
		return nil, err
	} else if pin != nil {
		log.Printf("worker.poll.found pin_id=%s", pin.Id)
		return pin, nil
	}
	return nil, nil
}

func workerQuery(p *pin) error {
	log.Printf("worker.query.start pin_id=%s", p.Id)
	pinDbUrl, err := dataPinDbUrl(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.reserve pin_id=%s", p.Id)
	startedAt := time.Now()
	p.QueryStartedAt = &startedAt
	err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.call pin_id=%s", p.Id)
	pinDb, err := sql.Open("postgres", pinDbUrl)
	if err != nil {
		return err
	}
	resultsRows, err := pinDb.Query(p.Query)
	defer resultsRows.Close()
	if err != nil {
		return err
	}
	log.Printf("worker.query.read pin_id=%s", p.Id)
	resultsFieldsData, err := resultsRows.Columns()
	if err != nil {
		return err
	}
	resultsRowsData := make([][]interface{}, 0)
	resultsRowsSeen := 0
	for resultsRows.Next() {
		resultsRowsSeen += 1
		if resultsRowsSeen > resultsRowsMax {
			return errors.New("too many rows")
		}
		resultsRowData := make([]interface{}, len(resultsFieldsData))
		resultsRowPointers := make([]interface{}, len(resultsFieldsData))
		for i, _ := range resultsRowData {
			resultsRowPointers[i] = &resultsRowData[i]
		}
		err := resultsRows.Scan(resultsRowPointers...)
		if err != nil {
			return err
		}
		resultsRowsData = append(resultsRowsData, resultsRowData)
	}
	err = resultsRows.Err()
	if err != nil {
		return nil
	}
	p.ResultsFields = MustNewPgJson(resultsFieldsData)
	p.ResultsRows = MustNewPgJson(resultsRowsData)
	finishedAt := time.Now()
	p.QueryFinishedAt = &finishedAt
	log.Printf("worker.query.commit pin_id=%s", p.Id)
	err = dataPinUpdate(p)
	if err != nil {
		return err
	}
	log.Printf("worker.query.finish pin_id=%s", p.Id)
	return nil
}

func workerTick() {
	pin, err := workerPoll()
	if err != nil {
		panic(err)
	}
	if pin != nil {
		err = workerQuery(pin)
		if err != nil {
			panic(err)
		}
	} else {
		time.Sleep(time.Millisecond * 250)
	}
}

func workerStart() {
	log.Print("worker.start")
	for {
		workerTick()
	}
}
