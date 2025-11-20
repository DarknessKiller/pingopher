package notification

import (
	"net/http"

	"github.com/DarknessKiller/pingopher/internal/dto"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
	"github.com/DarknessKiller/pingopher/internal/util"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *notification.NotificationService
}

func New(service *notification.NotificationService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateNotification(ctx *gin.Context) {
	hostID := ctx.Param("id")
	notification, err := dto.BindAndMap[dto.CreateNotificationRequest](ctx)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	err = h.service.CreateNotification(ctx, hostID, notification)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "notification created"})
}

func (h *Handler) GetAllNotificationsForHost(ctx *gin.Context) {
	hostID := ctx.Param("id")
	notifications, err := h.service.GetNotificationsForHost(ctx, hostID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.ToNotifications(notifications))
}

func (h *Handler) DeleteNotification(ctx *gin.Context) {
	notificationID := ctx.Param("id")

	err := h.service.DeleteNotification(ctx, notificationID)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Handler) UpdateNotification(ctx *gin.Context) {
	notificationID := ctx.Param("id")
	notification, err := dto.BindAndMap[dto.UpdateNotificationRequest](ctx)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	err = h.service.UpdateNotification(ctx, notificationID, notification)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "notification updated"})
}
