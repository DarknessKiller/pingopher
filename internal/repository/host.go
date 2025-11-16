package repository

import (
	"github.com/DarknessKiller/pingopher/internal/model"
)

type HostRepository interface {
	Repository[model.Host]
}
