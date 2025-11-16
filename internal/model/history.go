package model

import (
	"database/sql"

	"github.com/segmentio/ksuid"
)

type History struct {
	BaseModel
	HostID       ksuid.KSUID  `gorm:"type:varchar(27);index"`
	Host         Host         `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
	StatusCode   uint16       `gorm:"type:smallint unsigned;not null"`
	Latency      uint16       `gorm:"type:smallint unsigned;not null"`
	PingDateTime sql.NullTime `gorm:"type:timestamp;not null"`
	DNS          DNS          `gorm:"type:json"`
	ErrorMessage string
	Timestamps
}
