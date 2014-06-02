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
		first := args[0].(string)
		rest := args[1:]
		line = fmt.Sprintf(k+" "+first, rest...)
	}
	logMutex.Lock()
	defer logMutex.Unlock()
	fmt.Println(line)
}
