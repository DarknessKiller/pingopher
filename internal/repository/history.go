package repository

import (
	"context"

	"github.com/DarknessKiller/pingopher/internal/model"
	"gorm.io/gorm"
)

type HistoryRepository interface {
	Repository[model.History]
	GetHistoryByID(ctx context.Context, historyId string) (*model.History, error)
	CreatePingHistory(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error)
}

type historyRepository struct {
	*BaseRepository[model.History]
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepository{
		BaseRepository: NewBaseRepository[model.History](db),
	}
}

func (h *historyRepository) GetHistoryByID(ctx context.Context, historyId string) (*model.History, error) {
	var history model.History
	err := h.db.WithContext(ctx).Preload("Host").First(&history, "`id` = ?", historyId).Error
	return &history, err
}

func (h *historyRepository) CreatePingHistory(
	ctx context.Context,
	host *model.Host,
	histories []*model.History,
) ([]*model.History, error) {

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(histories).Error; err != nil {
			return err

		}
		if err := tx.Model(host).Updates(host).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return histories, err
}
