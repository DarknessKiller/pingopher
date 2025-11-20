package handler

import (
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/gin-gonic/gin"
)

var migrations = []interface{}{
	&model.Host{},
	&model.History{},
	&model.Notification{},
}

func (h *Handler) Migration(ctx *gin.Context) {

	if err := h.db.AutoMigrate(migrations...); err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"status": "ok", "message": "database migrated"})
}
