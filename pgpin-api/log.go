package main

import (
	"fmt"
	"sync"
)

var logMutex *sync.Mutex

func logInit() {
	logMutex = &sync.Mutex{}
}

func log(s string, args ...interface{}) {
	line := fmt.Sprintf(s, args...)
	logMutex.Lock()
	defer logMutex.Unlock()
	fmt.Println(line)
}

