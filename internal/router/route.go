package router

import (
	"net/http"

	"github.com/DarknessKiller/pingopher/internal/bootstrap"
	"github.com/DarknessKiller/pingopher/internal/handler"
	notification_handler "github.com/DarknessKiller/pingopher/internal/handler/notification"
	uptime_handler "github.com/DarknessKiller/pingopher/internal/handler/uptime"
	"github.com/gin-gonic/gin"
)

func Router(bs *bootstrap.Bootstrap) *gin.Engine {

	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	handler := handler.New(bs)
	uptimeHandler := uptime_handler.New(bs.UptimeService, bs.UptimeScheduler)
	notificationHandler := notification_handler.New(bs.NotificationService)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.Health)
		v1.POST("/migration", handler.Migration)

		uptime := v1.Group("/uptime")
		{
			uptime.POST("/create", uptimeHandler.CreateHost)
			uptime.GET("/all", uptimeHandler.GetAllHosts)
			hostGroup := uptime.Group("/:id")
			{
				hostGroup.PUT("", uptimeHandler.UpdateHost)
				hostGroup.DELETE("", uptimeHandler.DeleteHost)
				hostGroup.GET("", uptimeHandler.PingHost)
				hostGroup.GET("/history", uptimeHandler.GetHistory)

				notification := hostGroup.Group("/notification")
				{
					notification.POST("", notificationHandler.CreateNotification)
					notification.GET("", notificationHandler.GetAllNotificationsForHost)
					notification.DELETE("/:id", notificationHandler.DeleteNotification)
					notification.PUT(":id", notificationHandler.UpdateNotification)
				}
			}
		}
	}
	return r
}
