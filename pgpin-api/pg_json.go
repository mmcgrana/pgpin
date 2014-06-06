package main

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

type NullJson struct {
	Json  []byte
	Valid bool
}

var nullValue = []byte(`null`)

func (p *NullJson) MarshalJSON() ([]byte, error) {
	if p.Valid {
		return p.Json, nil
	}
	return nullValue, nil
}

func (p *NullJson) UnmarshalJSON(data []byte) error {
	if p == nil {
		return errors.New("pg_json: UnmarshalJSON on nil pointer")
	}
	if bytes.Equal(nullValue, data) {
		p.Valid = false
		return nil
	}
	p.Valid = true
	json := make([]byte, len(data))
	copy(json, data)
	p.Json = json
	return nil
}

func (p *NullJson) Scan(value interface{}) error {
	p.Json, p.Valid = value.([]byte)
	return nil
}

func (p NullJson) Value() (driver.Value, error) {
	if !p.Valid {
		return nil, nil
	}
	return p.Json, nil
}
