package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
)

type stringMap map[string]string

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing " + k)
	}
	return v
}

func mustRandRead(b []byte) {
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		panic("failed to rand.Read")
	}
}

func randBytes(l int) []byte {
	b := make([]byte, l)
	mustRandRead(b)
	return b
}

func randId() string {
	uuid := randBytes(6)
	return fmt.Sprintf("%x", uuid)
}

func confPort() int {
	port, err := strconv.Atoi(mustGetenv("PORT"))
	if err != nil {
		panic(err)
	}
	return port
}
