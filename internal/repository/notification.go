package repository

import (
	"context"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Repository[model.Notification]
	GetActiveNotificationsForHost(ctx context.Context, hostID string) ([]model.Notification, error)
	UpdateLastNotifiedAt(ctx context.Context, id string) error
}

type notificationRepository struct {
	*BaseRepository[model.Notification]
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		BaseRepository: NewBaseRepository[model.Notification](db),
	}
}

func (r *notificationRepository) GetActiveNotificationsForHost(ctx context.Context, hostID string) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.WithContext(ctx).Where("`host_id` = ? AND `active` = ?", hostID, true).Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) UpdateLastNotifiedAt(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("`id` = ?", id).Update("last_notified_at", time.Now()).Error
}

func (r *notificationRepository) GetNotificationsForHost(ctx context.Context, hostID string) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.WithContext(ctx).Where("`host_id` = ?", hostID).Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}
