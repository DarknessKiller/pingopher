package uptime

import (
	"github.com/DarknessKiller/pingopher/internal/repository"
)

type Repository interface {
	Host() repository.HostRepository
	History() repository.HistoryRepository
}

type newRepository struct {
	HostRepo    repository.HostRepository
	HistoryRepo repository.HistoryRepository
}

func (r *newRepository) Host() repository.HostRepository {
	return r.HostRepo
}

func (r *newRepository) History() repository.HistoryRepository {
	return r.HistoryRepo
}
