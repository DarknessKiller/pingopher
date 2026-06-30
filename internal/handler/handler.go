package handler

import (
	"github.com/DarknessKiller/pingopher/internal/bootstrap"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func New(bs *bootstrap.Bootstrap) *Handler {
	return &Handler{db: bs.Database}
}
