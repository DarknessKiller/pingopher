package model

import (
	"database/sql"

	"github.com/segmentio/ksuid"
)

type History struct {
	BaseModel
	HostID       ksuid.KSUID  `gorm:"type:varchar(27);index:idx_history_host_datetime,priority:1"`
	Host         Host         `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
	StatusCode   uint16       `gorm:"type:smallint unsigned;not null;index:idx_history_error,priority:1"`
	Latency      uint16       `gorm:"type:smallint unsigned;not null"`
	PingDateTime sql.NullTime `gorm:"type:timestamp;index:idx_history_host_datetime,priority:2"`
	DNS          DNS          `gorm:"type:json"`
	ErrorMessage string       `gorm:"index:idx_history_error,priority:2"`
	Timestamps
}
