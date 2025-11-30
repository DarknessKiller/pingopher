package repository

import (
	"context"
	"strings"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/util"
	"gorm.io/gorm"
)

type HistoryRepository interface {
	Repository[model.History]
	GetHistoryByID(ctx context.Context, historyId string) (*model.History, error)
	GetHistoryByHostID(ctx context.Context, hostID string, startAt, endAt time.Time, status *[]string) ([]*model.History, error)
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

func (r *historyRepository) GetHistoryByHostID(
	ctx context.Context,
	hostID string,
	startAt, endAt time.Time,
	status *[]string,
) ([]*model.History, error) {
	var histories []*model.History

	query := r.db.WithContext(ctx).
		Where("host_id = ?", hostID).
		Where("ping_date_time BETWEEN ? AND ?", startAt, endAt).
		Order("ping_date_time ASC")

	if status != nil && len(*status) > 0 {
		statusRange, err := util.ParseStatusCode(*status)
		if err != nil {
			return nil, err
		}

		if len(statusRange) > 0 {
			whereClause, args := buildStatusWhereClause(statusRange)
			query = query.Where(whereClause, args...)
		}
	}

	err := query.Find(&histories).Error
	if err != nil {
		return nil, err
	}
	return histories, nil
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

func buildStatusWhereClause(ranges []util.StatusCodeRange) (string, []interface{}) {
	var parts []string
	var args []interface{}

	for _, r := range ranges {
		if r.From == r.To {
			parts = append(parts, "status_code = ?")
			args = append(args, r.From)
		} else {
			parts = append(parts, "(status_code BETWEEN ? AND ?)")
			args = append(args, r.From, r.To)
		}
	}

	return strings.Join(parts, " OR "), args
}
