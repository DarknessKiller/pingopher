package uptime_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/DarknessKiller/pingopher/internal/dto"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
	uptime "github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/DarknessKiller/pingopher/internal/util"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service      *uptime.Service
	notification *notification.NotificationService
	scheduler    *uptime.Scheduler
}

func New(service *uptime.Service, notification *notification.NotificationService, scheduler *uptime.Scheduler) *Handler {
	return &Handler{service: service, notification: notification, scheduler: scheduler}
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

	go func(ctx context.Context, hostID string) {
		ctx = context.WithoutCancel(ctx)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		_, _, histories, err := h.service.PingHost(ctx, hostID)
		if err != nil {
			return
		}

		h.notification.SendNotification(host, histories)
	}(ctx.Request.Context(), host.ID.String())

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
	hostID := ctx.Param("hostId")

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

	go func(ctx context.Context, hostID string) {
		ctx = context.WithoutCancel(ctx)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		_, _, histories, err := h.service.PingHost(ctx, hostID)
		if err != nil {
		}

		h.notification.SendNotification(host, histories)
	}(ctx.Request.Context(), host.ID.String())

	ctx.JSON(200, dto.ToHost(host))
}

func (h *Handler) DeleteHostAndNotifications(ctx *gin.Context) {
	hostID := ctx.Param("hostId")

	err := h.notification.DeleteNotifications(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	err = h.service.DeleteHost(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	h.scheduler.DeleteHost(ctx, hostID)

	ctx.Status(204)
}

func (h *Handler) PingHost(ctx *gin.Context) {
	hostID := ctx.Param("hostId")

	_, _, histories, err := h.service.PingHost(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))
}

func (h *Handler) GetHistory(ctx *gin.Context) {
	hostID := ctx.Param("hostId")
	startAt := ctx.Query("startAt")
	endAt := ctx.Query("endAt")

	if startAt == "" || endAt == "" {
		util.HandleError(ctx, fmt.Errorf("startAt and endAt parameters are required"))
		return
	}

	host, err := h.service.GetHostByID(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	histories, err := h.service.GetHistoryByHostID(ctx, host.ID.String(), startAt, endAt)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(200, dto.ToHistories(histories))
}
