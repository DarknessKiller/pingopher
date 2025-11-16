package model

import (
	"database/sql/driver"
	"encoding/json"
)

type DNS struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     uint16 `json:"port"`
	Protocol string `json:"protocol"`
}

func (d DNS) Value() (driver.Value, error) {
	switch d {
	case DNS{}:
		return nil, nil
	default:
		return json.Marshal(d)
	}
}

func (d *DNS) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	default:
		return nil
	}
}

type DNSs []DNS

func (d DNSs) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	return json.Marshal(d)
}

func (d *DNSs) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	default:
		return nil
	}
}
