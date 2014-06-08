package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type PgJson []byte

var nullValue = []byte(`null`)

func (p *PgJson) MarshalJSON() ([]byte, error) {
	if len(*p) == 0 {
		return nullValue, nil
	}
	return []byte(*p), nil
}

func (p *PgJson) UnmarshalJSON(data []byte) error {
	if p == nil {
		return errors.New("pg_json: UnmarshalJSON on nil pointer")
	}
	json := make([]byte, len(data))
	copy(json, data)
	*p = json
	return nil
}

// Scan updates the called PgJson struct to contain valid
// JSON according to the given value, which we expect to be
// nil or a []byte of valid JSON.
func (p *PgJson) Scan(value interface{}) error {
	if value == nil {
		*p = nullValue
		return nil
	}
	*p = PgJson(value.([]byte))
	return nil
}

// Value returns a []byte representation of the called
// PgJson struct.
func (p PgJson) Value() (driver.Value, error) {
	return []byte(p), nil
}

// MustNewPgJson returns a PgJson struct corresponding to
// the given data, which shuold be a data structure (not an
// allready-encoded JSON string).
func MustNewPgJson(data interface{}) PgJson {
	encoded, err := json.Marshal(data)
	must(err)
	return PgJson(encoded)
}
