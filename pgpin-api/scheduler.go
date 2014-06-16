package main

import (
	"fmt"
	"github.com/jrallison/go-workers"
	"log"
	"time"
)

func SchedulerEnqueue(pin *Pin) error {
	log.Printf("scheduler.enqueue pin_id=%s", pin.Id)
	pin.ScheduledAt = time.Now()
	err := DataPinUpdate(pin)
	if err != nil {
		return err
	}
	err = workers.Enqueue("pins", "", pin.Id)
	if err != nil {
		return err
	}
	return nil
}

func SchedulerError(err error) {
	log.Printf("scheduler.error %+s", err.Error())
}

func SchedulerStart() {
	log.Printf("scheduler.start")
	DataStart()
	QueueStart()
	for {
		log.Printf("scheduler.tick")
		queryFrag := fmt.Sprintf("scheduled_at < now()-'%f seconds'::interval", ConfigPinRefreshInterval.Seconds())
		ready, err := DataPinList(queryFrag)
		if err != nil {
			SchedulerError(err)
		} else {
			for _, pin := range ready {
				err = SchedulerEnqueue(pin)
				if err != nil {
					SchedulerError(err)
				}
			}
		}
		time.Sleep(ConfigSchedulerTickInterval)
	}
}
