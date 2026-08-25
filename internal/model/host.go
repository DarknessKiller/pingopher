package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/lib/pq"
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
	Name                string         `gorm:"type:varchar(255);not null;index:idx_host_name"`
	Protocol            string         `gorm:"type:varchar(6);not null"`
	HostURL             string         `gorm:"type:varchar(255);not null;index:uniq_host_url_port,priority:1"`
	Port                *uint16        `gorm:"type:smallint unsigned;index:uniq_host_url_port,priority:2"`
	TLS                 TLS            `gorm:"type:json"`
	DNS                 DNSs           `gorm:"type:json"`
	PingInterval        uint16         `gorm:"type:smallint unsigned;not null"`
	FailThreshold       uint16         `gorm:"type:smallint unsigned;not null"`
	AcceptedStatusCodes pq.StringArray `gorm:"type:varchar(255)"`
	Status              HostStatus     `gorm:"type:varchar(8);not null;index:idx_host_status"`
	Timestamps
	DeletedAt gorm.DeletedAt `gorm:"index"`

	History []History `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
}

// DisplayURL renders a human-readable target for the host's protocol:
// http(s)://host[:port], host:port for tcp/udp, and just host for ping.
func (h Host) DisplayURL() string {
	hostPort := h.HostURL
	if h.Port != nil && *h.Port != 0 {
		hostPort = net.JoinHostPort(h.HostURL, fmt.Sprintf("%d", *h.Port))
	}
	switch h.Protocol {
	case "tcp", "udp":
		return hostPort
	case "ping":
		return h.HostURL
	default:
		return h.Protocol + "://" + hostPort
	}
}

type TLS struct {
	NoVerify *bool `json:"no_verify"`
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
