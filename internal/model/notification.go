package model

import (
	"database/sql"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	DiscordNotification NotificationType = "discord"
)

type Notification struct {
	BaseModel
	HostID         ksuid.KSUID      `gorm:"type:char(27);index:idx_notification_host,priority:1;not null"`
	Host           Host             `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
	Name           string           `gorm:"type:varchar(255);not null"`
	Type           NotificationType `gorm:"type:varchar(50);index:idx_notification_host,priority:2;not null"`
	Active         *bool            `gorm:"type:boolean;not null"`
	LastNotifiedAt sql.NullTime     `gorm:"index"`

	// === Discord-specific fields  ===
	DiscordUsername      string `gorm:"type:varchar(100)"`
	DiscordWebhookURL    string `gorm:"type:text"`
	DiscordPrefixMessage string `gorm:"type:text"`
	DiscordDisableURL    bool
	DiscordChannelType   string `gorm:"type:varchar(50)"`
	DiscordThreadID      string `gorm:"type:varchar(100)"`
	DiscordPostName      string `gorm:"type:varchar(255)"`

	Timestamps
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
