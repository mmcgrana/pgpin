package main

import (
	"fmt"
	"regexp"
)

var EmptyRegexp = regexp.MustCompile("\\A\\s*\\z")

func ValidateNonempty(f string, s string) error {
	if EmptyRegexp.MatchString(s) {
		return &PgpinError{
			Id:         "invalid",
			Message:    fmt.Sprintf("field %s must be nonempty", f),
			HttpStatus: 400,
		}
	}
	return nil
}

var SlugRegexp = regexp.MustCompile("\\A[a-z0-9-]+\\z")

func ValidateSlug(f string, s string) error {
	if !SlugRegexp.MatchString(s) {
		return &PgpinError{
			Id:         "invalid",
			Message:    fmt.Sprintf("field %s must be of the form [a-z0-9-]+", f),
			HttpStatus: 400,
		}
	}
	return nil
}

func ValidatePgUrl(f string, s string) error {
	return ValidateNonempty(f, s)
}
