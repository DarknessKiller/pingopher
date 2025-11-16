package model

import (
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

type Models interface {
	Host | History | User
}

type BaseModel struct {
	ID ksuid.KSUID `gorm:"primaryKey;type:varchar(27);not null"`
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {

	m.ID = ksuid.New()
	return nil
}
