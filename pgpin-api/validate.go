package main

import (
	"fmt"
	"regexp"
)

var emptyRegexp = regexp.MustCompile("\\A\\s*\\z")

func validateNonempty(f string, s string) error {
	if emptyRegexp.MatchString(s) {
		return &pgpinError{
			Id:         "invalid",
			Message:    fmt.Sprintf("field %s must be nonempty", f),
			HttpStatus: 400,
		}
	}
	return nil
}

var slugRegexp = regexp.MustCompile("\\A[a-z0-9-]+\\z")

func validateSlug(f string, s string) error {
	if !slugRegexp.MatchString(s) {
		return &pgpinError{
			Id:         "invalid",
			Message:    fmt.Sprintf("field %s must be of the form [a-z0-9-]+", f),
			HttpStatus: 400,
		}
	}
	return nil
}

func validatePgUrl(f string, s string) error {
	return validateNonempty(f, s)
}
