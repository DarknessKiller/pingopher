package notification

import (
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/repository"
)

type Repository interface {
	Host() *repository.BaseRepository[model.Host]
	Notification() repository.NotificationRepository
}

type newRepository struct {
	HostRepo         *repository.BaseRepository[model.Host]
	NotificationRepo repository.NotificationRepository
}

func (r *newRepository) Host() *repository.BaseRepository[model.Host] {
	return r.HostRepo
}

func (r *newRepository) Notification() repository.NotificationRepository {
	return r.NotificationRepo
}
