package main

import (
	"fmt"
	"log"
	"os"
)

func init() {
	log.SetFlags(0)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: datapins-api [web|worker]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "web":
		dataStart()
		webStart()
	case "worker":
		dataStart()
		workerStart()
	default:
		usage()
	}
}
