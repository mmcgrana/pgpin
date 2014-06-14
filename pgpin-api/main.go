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
	_, err := fmt.Fprintln(os.Stderr, "Usage: datapins-api [web|worker]")
	Must(err)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "web":
		WebStart()
	case "worker":
		WorkerStart()
	default:
		usage()
	}
}
