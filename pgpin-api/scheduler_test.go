package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerNoEnqueues(t *testing.T) {
	defer clear()
	dbIn := mustDbAdd("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustPinGet(pinIn.Id)
	mustSchedulerTick()
	mustWorkerTick()
	pinOut2 := mustPinGet(pinIn.Id)
	assert.Equal(t, pinOut1.Version, pinOut2.Version)
}

func TestSchedulerEnqueue(t *testing.T) {
	defer clear()
	dbIn := mustDbAdd("dbs-1", ConfigDatabaseUrl)
	pinIn := mustPinCreate(dbIn.Id, "pins-1", "select now()")
	mustWorkerTick()
	pinOut1 := mustPinGet(pinIn.Id)
	ConfigPinRefreshIntervalPrev := ConfigPinRefreshInterval
	defer func() {
		ConfigPinRefreshInterval = ConfigPinRefreshIntervalPrev
	}()
	ConfigPinRefreshInterval = 0
	mustSchedulerTick()
	mustWorkerTick()
	pinOut2 := mustPinGet(pinIn.Id)
	assert.NotEqual(t, pinOut1.Version, pinOut2.Version)
}
