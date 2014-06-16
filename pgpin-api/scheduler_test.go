package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerNoEnqueues(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", ConfigDatabaseUrl)
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustDataPinGet(pinIn.Id)
	mustSchedulerTick()
	mustWorkerTick()
	pinOut2 := mustDataPinGet(pinIn.Id)
	assert.Equal(t, pinOut1.Version, pinOut2.Version)
}

func TestSchedulerEnqueue(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", ConfigDatabaseUrl)
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustDataPinGet(pinIn.Id)
	ConfigPinRefreshIntervalPrev := ConfigPinRefreshInterval
	defer func() {
		ConfigPinRefreshInterval = ConfigPinRefreshIntervalPrev
	}()
	ConfigPinRefreshInterval = 0
	mustSchedulerTick()
	mustWorkerTick()
	pinOut2 := mustDataPinGet(pinIn.Id)
	assert.NotEqual(t, pinOut1.Version, pinOut2.Version)
}
