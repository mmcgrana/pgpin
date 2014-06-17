package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// PgJson represents JSON values stored in Postgres using the json
// data type.
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
	bytes := value.([]byte)
	bytesCopy := make([]byte, len(bytes))
	copy(bytesCopy, bytes)
	*p = PgJson(bytesCopy)
	return nil
}

// Value returns a []byte representation of the called
// PgJson struct.
func (p PgJson) Value() (driver.Value, error) {
	return []byte(p), nil
}

// MustNewPgJson returns a PgJson struct corresponding to
// the given data, which should be a data structure (not an
// already-encoded JSON string).
func MustNewPgJson(data interface{}) PgJson {
	encoded, err := json.Marshal(data)
	Must(err)
	return PgJson(encoded)
}
