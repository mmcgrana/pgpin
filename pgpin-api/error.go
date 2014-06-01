package main

import (
	"fmt"
)

type pgpinError struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

func (e pgpinError) Error() string {
	return fmt.Sprintf("pgpin: %s - %s", e.Id, e.Message)
}
