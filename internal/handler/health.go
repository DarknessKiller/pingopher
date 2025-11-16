package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "ok", "message": "Go backend is alive"})
}
