package main

import (
	"log"
)

func ScriptStart() {
    DataStart()
    count, _ := DataCount("SELECT count(*) from pins")
    log.Printf("pins.count total=%d", count)
}
