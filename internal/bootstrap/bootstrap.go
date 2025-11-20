package bootstrap

import (
	"log"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/repository"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
	"github.com/DarknessKiller/pingopher/internal/service/uptime"

	"gorm.io/gorm"
)

type Bootstrap struct {
	Config                 *config.Config
	Database               *gorm.DB
	HostRepository         *repository.BaseRepository[model.Host]
	HistoryRepository      repository.HistoryRepository
	NotificationRepository repository.NotificationRepository
	UptimeService          *uptime.Service
	NotificationService    *notification.NotificationService
	UptimeScheduler        *uptime.Scheduler
}

func New() *Bootstrap {
	bs := &Bootstrap{}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	bs.Config = cfg
	bs.Database = InitiateDatabase(bs.Config)
	bs.HostRepository = repository.NewBaseRepository[model.Host](bs.Database)
	bs.HistoryRepository = repository.NewHistoryRepository(bs.Database)
	bs.NotificationRepository = repository.NewNotificationRepository(bs.Database)
	bs.UptimeService = uptime.NewService(bs.Config, bs.HostRepository, bs.HistoryRepository)
	bs.NotificationService = notification.NewService(bs.HostRepository, bs.NotificationRepository)
	bs.UptimeScheduler = uptime.NewScheduler(bs.UptimeService, bs.NotificationService)
	return bs
}
