package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type HostStatus string

const (
	HostStatusUnknown HostStatus = "unknown"
	HostStatusUp      HostStatus = "up"
	HostStatusDown    HostStatus = "down"
)

type Host struct {
	BaseModel
	Name          string     `gorm:"type:varchar(255);not null;index:idx_host_name"`
	Protocol      string     `gorm:"type:varchar(6);not null"`
	HostURL       string     `gorm:"type:varchar(255);not null;index:uniq_host_url_port,priority:1"`
	Port          uint16     `gorm:"type:smallint unsigned;not null;index:uniq_host_url_port,priority:2"`
	TLS           TLS        `gorm:"type:json"`
	DNS           DNSs       `gorm:"type:json"`
	PingInterval  uint16     `gorm:"type:smallint unsigned;not null"`
	FailThreshold uint16     `gorm:"type:smallint unsigned;not null"`
	Status        HostStatus `gorm:"type:varchar(8);not null;index:idx_host_status"`
	Timestamps
	DeletedAt gorm.DeletedAt `gorm:"index"`

	History []History `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
}

type TLS struct {
	NoVerify bool `json:"no_verify"`
}

func (t TLS) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TLS) Scan(value interface{}) error {
	if value == nil {
		*t = TLS{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	default:
		return errors.New("scan: expected []byte or string")
	}
}
