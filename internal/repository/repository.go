package repository

import (
	"context"

	"github.com/DarknessKiller/pingopher/internal/model"
	"gorm.io/gorm"
)

type Repository[T model.Models] interface {
	Create(ctx context.Context, model *T) error
	GetByID(ctx context.Context, id string) (*T, error)
	GetAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, id string, model *T) error
	Delete(ctx context.Context, id string) error
}

type BaseRepository[T model.Models] struct {
	db *gorm.DB
}

func NewBaseRepository[T model.Models](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var model T

	err := r.db.WithContext(ctx).Where("`id` =  ?", id).First(&model).Error
	return &model, err
}

func (r *BaseRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

func (r *BaseRepository[T]) Update(ctx context.Context, id string, model *T) error {
	if err := r.db.WithContext(ctx).Model(&model).
		Where("`id` = ?", id).
		Updates(model).Error; err != nil {
		return err
	}

	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	return r.db.WithContext(ctx).Where("`id` = ?", id).Delete(&entity).Error
}
