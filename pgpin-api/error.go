package main

import (
	"fmt"
)

type pgpinError struct {
	Id         string `json:"id"`
	Message    string `json:"message"`
	HttpStatus int    `json:"-"`
}

func (e pgpinError) Error() string {
	return fmt.Sprintf("pgpin: %s - %s", e.Id, e.Message)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
