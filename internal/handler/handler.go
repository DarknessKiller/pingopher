package handler

import (
	"github.com/DarknessKiller/pingopher/internal/bootstrap"
	"github.com/DarknessKiller/pingopher/internal/config"
	"gorm.io/gorm"
)

type Handler struct {
	config *config.Config
	db     *gorm.DB
}

func New(bs *bootstrap.Bootstrap) *Handler {
	return &Handler{config: bs.Config, db: bs.Database}
}
