package model

import "gorm.io/gorm"

type Host struct {
	BaseModel
	Name          string `gorm:"type:varchar(255);not null"`
	Protocol      string `gorm:"type:varchar(6);not null"`
	HostURL       string `gorm:"type:varchar(255);not null"`
	Port          uint16 `gorm:"type:smallint unsigned;not null"`
	DNS           DNSs   `gorm:"type:json"`
	PingInterval  uint16 `gorm:"type:smallint unsigned;not null"`
	FailThreshold uint16 `gorm:"type:smallint unsigned;not null"`
	Status        string `gorm:"default:unknown"`
	Timestamps
	DeletedAt gorm.DeletedAt

	History []History `gorm:"foreignKey:HostID;references:ID;-:migration;->"`
}
