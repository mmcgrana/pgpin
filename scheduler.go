package main

import (
	"github.com/jrallison/go-workers"
	"log"
	"time"
)

func SchedulerEnqueue(pin *Pin) error {
	log.Printf("scheduler.enqueue pin_id=%s", pin.Id)
	pin.ScheduledAt = time.Now()
	err := PinUpdate(pin)
	if err != nil {
		return err
	}
	err = workers.Enqueue("pins", "", pin.Id)
	if err != nil {
		return err
	}
	return nil
}

func SchedulerTick() error {
	log.Printf("scheduler.tick")
	ready, err := PinList("scheduled_at <= $1", time.Now().Add(-ConfigPinRefreshInterval))
	if err != nil {
		return err
	}
	for _, pin := range ready {
		err = SchedulerEnqueue(pin)
		if err != nil {
			return err
		}
	}
	return nil
}

func SchedulerStart() {
	log.Printf("scheduler.start")
	PgStart()
	RedisStart()
	for {
		err := SchedulerTick()
		if err != nil {
			log.Printf("scheduler.error %+s", err.Error())
		}
		time.Sleep(ConfigSchedulerTickInterval)
	}
}
