package util

import (
	"errors"

	"github.com/DarknessKiller/pingopher/internal/model"
)

func SendNotificationWebhook(host *model.Host, histories []*model.History) error {
	return errors.New("notification webhook not implemented")
}
