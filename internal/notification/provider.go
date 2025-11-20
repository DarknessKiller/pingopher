package notification

import "github.com/DarknessKiller/pingopher/internal/model"

type Provider interface {
	Send(host model.Host, histories []*model.History) error
}
