package uptime_handler

import (
	"fmt"

	"github.com/DarknessKiller/pingopher/internal/dto"
	uptime "github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/DarknessKiller/pingopher/internal/util"
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
		util.HandleError(ctx, err)
		return
	}

	err = h.service.CreateHost(ctx, host)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	h.scheduler.ScheduleHost(ctx, host)

	ctx.JSON(201, dto.ToHost(host))
}

func (h *Handler) GetAllHosts(ctx *gin.Context) {
	hosts, err := h.service.GetAllHosts(ctx)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToAllHosts(hosts))
}

func (h *Handler) UpdateHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	host, err := dto.BindAndMap[dto.UpdateHostRequest](ctx)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	host, err = h.service.UpdateHost(ctx, hostID, host)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	h.scheduler.ScheduleHost(ctx, host)

	ctx.JSON(200, dto.ToHost(host))
}

func (h *Handler) DeleteHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	err := h.service.DeleteHost(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	h.scheduler.DeleteHost(ctx, hostID)

	ctx.Status(204)
}

func (h *Handler) PingHost(ctx *gin.Context) {
	hostID := ctx.Param("id")

	_, _, histories, err := h.service.PingHost(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))
}

func (h *Handler) GetHistory(ctx *gin.Context) {
	hostID := ctx.Param("id")
	startAt := ctx.Query("startAt")
	endAt := ctx.Query("endAt")

	if startAt == "" || endAt == "" {
		util.HandleError(ctx, fmt.Errorf("startAt and endAt parameters are required"))
		return
	}

	histories, err := h.service.GetHistoryByHostID(ctx, hostID, startAt, endAt)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))
}
