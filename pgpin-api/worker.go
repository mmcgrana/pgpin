package main

import (
	"bitbucket.org/kardianos/tablebuffer"
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func workerPoll() (*pin, error) {
	pin, err := dataGetPinInternal("WHERE query_started_at IS NULL AND deleted_at IS NULL")
	if err != nil {
		return nil, err
	} else if pin != nil {
		log("key=worker.poll.found pin_id=%s", pin.Id)
		return pin, nil
	}
	return nil, nil
}

func workerQuery(p *pin) error {
	log("key=worker.query.start pin_id=%s", p.Id)
	resourceConf, err := pq.ParseURL(*p.ResourceUrl)
	if err != nil {
		return err
	}
	log("key=worker.query.reserve pin_id=%s", p.Id)
	startedAt := time.Now()
	p.QueryStartedAt = &startedAt
	err = dataUpdatePin(p)
	if err != nil {
		return err
	}
	log("key=worker.query.open pin_id=%s", p.Id)
	resourceDb, err := sql.Open("postgres", resourceConf)
	if err != nil {
		return err
	}
	log("key=worker.query.exec pin_id=%s", p.Id)
	buffer, err := tablebuffer.Get(resourceDb, p.Sql)
	finishedAt := time.Now()
	p.QueryFinishedAt = &finishedAt
	p.ResourceUrl = nil
	if err != nil {
		if pgerr, ok := err.(pq.PGError); ok {
			log("key=worker.query.usererror pin_id=%s", p.Id)
			msg := pgerr.Get('M')
			p.ErrorMessage = &msg
			err = dataUpdatePin(p)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	log("key=worker.query.read pin_id=%s", p.Id)
	resultsFieldsJsonB, _ := json.Marshal(buffer.ColumnNames)
	resultsFieldsJson := string(resultsFieldsJsonB)
	resultsRows := make([][]interface{}, len(buffer.Rows))
	for i, row := range buffer.Rows {
		resultsRows[i] = make([]interface{}, len(row.Data))
		for j, rowDatum := range row.Data {
			switch rowValue := rowDatum.(type) {
			case []byte:
				resultsRows[i][j] = string(rowValue)
			default:
				resultsRows[i][j] = rowValue
			}
		}
	}
	resultsRowsJsonB, _ := json.Marshal(resultsRows)
	resultsRowsJson := string(resultsRowsJsonB)
	log("key=worker.query.commit pin_id=%s", p.Id)
	p.ResultsFieldsJson = &resultsFieldsJson
	p.ResultsRowsJson = &resultsRowsJson
	err = dataUpdatePin(p)
	if err != nil {
		return err
	}
	log("key=worker.query.finish pin_id=%s", p.Id)
	return nil
}

func workerTrap() chan os.Signal {
	traps := make(chan os.Signal, 1)
	sigs := make(chan os.Signal, 1)
	go func() {
		s := <- traps
		log("key=worker.trap")
		sigs <- s
	}()
	signal.Notify(traps, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

func workerTrapped(sigs chan os.Signal) bool {
	select {
	case <-sigs:
		return true
	default:
	}
	return false
}

func worker() {
	dataInit()
	log("key=worker.start")
	sigs := workerTrap()
	for {
		if workerTrapped(sigs) {
			log("key=worker.exit")
			return
		}
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
}
