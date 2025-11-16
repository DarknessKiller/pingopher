package uptime_handler

import (
	"github.com/DarknessKiller/pingopher/internal/dto"
	uptime "github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *uptime.Service
}

func New(service *uptime.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateHost(ctx *gin.Context) {
	host, err := dto.BindAndMap[dto.CreateHostRequest](ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err = h.service.CreateHost(ctx, host)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(201, dto.ToHost(host))
}

func (h *Handler) GetAllHosts(ctx *gin.Context) {
	hosts, err := h.service.GetAllHosts(ctx)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(200, dto.ToAllHosts(hosts))
}

func (h *Handler) UpdateHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	host, err := dto.BindAndMap[dto.UpdateHostRequest](ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	host, err = h.service.UpdateHost(ctx, hostID, host)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(200, dto.ToHost(host))
}

func (h *Handler) DeleteHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	err := h.service.DeleteHost(ctx, hostID)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.Status(204)
}

func (h *Handler) PingHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	histories, _, err := h.service.PingHost(ctx, hostID)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))

}
