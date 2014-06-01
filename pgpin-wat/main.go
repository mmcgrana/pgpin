package main

import (
	"fmt"
	"os"
)

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
		web()
	case "worker":
		worker()
	case "scratch":
		scratch()
	}
}
