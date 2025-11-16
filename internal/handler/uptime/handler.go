package uptime_handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/DarknessKiller/pingopher/internal/dto"
	uptime "github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   *uptime.Service
	scheduler *uptime.Scheduler
}

func New(service *uptime.Service, scheduler *uptime.Scheduler) *Handler {
	return &Handler{service: service, scheduler: scheduler}
}

func (h *Handler) CreateHost(ctx *gin.Context) {
	host, err := dto.BindAndMap[dto.CreateHostRequest](ctx)
	if err != nil {
		handleError(ctx, err)
		return
	}

	err = h.service.CreateHost(ctx, host)
	if err != nil {
		handleError(ctx, err)
		return
	}
	h.scheduler.ScheduleHost(ctx, host)

	ctx.JSON(201, dto.ToHost(host))
}

func (h *Handler) GetAllHosts(ctx *gin.Context) {
	hosts, err := h.service.GetAllHosts(ctx)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToAllHosts(hosts))
}

func (h *Handler) UpdateHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	host, err := dto.BindAndMap[dto.UpdateHostRequest](ctx)
	if err != nil {
		handleError(ctx, err)
		return
	}

	host, err = h.service.UpdateHost(ctx, hostID, host)
	if err != nil {
		handleError(ctx, err)
		return
	}
	h.scheduler.ScheduleHost(ctx, host)

	ctx.JSON(200, dto.ToHost(host))
}

func (h *Handler) DeleteHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	err := h.service.DeleteHost(ctx, hostID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	h.scheduler.DeleteHost(ctx, hostID)

	ctx.Status(204)
}

func (h *Handler) PingHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	histories, _, err := h.service.PingHost(ctx, hostID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))
}

func handleError(ctx *gin.Context, err error) {
	returnError := func(status int, msg string) { ctx.JSON(status, gin.H{"status": "error", "message": msg}) }

	switch {
	case errors.As(err, &dto.ValidationError{}):
		returnError(http.StatusBadRequest, err.Error())
	case errors.Is(err, io.EOF):
	default:
		returnError(http.StatusInternalServerError, err.Error())
	}
}
