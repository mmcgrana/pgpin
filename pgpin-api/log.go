package main

import (
	"fmt"
	logger "log"
)

func init() {
	logger.SetFlags(0)
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
	logger.Println(line)
}
