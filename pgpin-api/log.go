package main

import (
	"fmt"
	"sync"
)

var logMutex *sync.Mutex

func init() {
	logMutex = &sync.Mutex{}
}

func log(k string, args ...interface{}) {
	var line string
	if len(args) == 0 {
		line = k
	} else {
		line = fmt.Sprintf(k + " " + args[0], args[1:]...)
	}
	logMutex.Lock()
	defer logMutex.Unlock()
	fmt.Println(line)
}
