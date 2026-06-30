package dto

import (
	"strings"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
)

type CreateNotificationRequest struct {
	Name   string `json:"name" binding:"required,max=100"`
	Type   string `json:"type" binding:"required,oneof=discord"`
	Active bool   `json:"active" binding:"required"`

	DiscordUsername      string `json:"discordUsername,omitempty"`
	DiscordWebhookURL    string `json:"discordWebhookUrl,omitempty" binding:"omitempty,url"`
	DiscordPrefixMessage string `json:"discordPrefixMessage,omitempty"`
	DiscordDisableURL    bool   `json:"discordDisableUrl,omitempty"`
	DiscordChannelType   string `json:"discordChannelType,omitempty" binding:"omitempty,oneof=postToThread createNewForumPost"`
	DiscordThreadID      string `json:"discordThreadId,omitempty" binding:"omitempty"`
	DiscordPostName      string `json:"discordPostName,omitempty" binding:"omitempty,max=100"`
}

func (r CreateNotificationRequest) ToModel() *model.Notification {

	switch strings.ToLower(r.Type) {
	case string(model.DiscordNotification):
		return &model.Notification{
			Name:                 r.Name,
			Type:                 model.DiscordNotification,
			Active:               &r.Active,
			DiscordWebhookURL:    r.DiscordWebhookURL,
			DiscordPrefixMessage: r.DiscordPrefixMessage,
			DiscordUsername:      r.DiscordUsername,
			DiscordDisableURL:    r.DiscordDisableURL,
			DiscordChannelType:   r.DiscordChannelType,
			DiscordThreadID:      r.DiscordThreadID,
			DiscordPostName:      r.DiscordPostName,
		}
	}

	return nil
}

type Notification struct {
	ID             string                 `json:"id"`
	HostID         string                 `json:"hostId"`
	Name           string                 `json:"name"`
	Type           model.NotificationType `json:"type"`
	Active         bool                   `json:"active"`
	LastNotifiedAt time.Time              `json:"lastNotifiedAt"`

	// === Discord-specific fields  ===
	DiscordUsername      string `json:"discordUsername"`
	DiscordWebhookURL    string `json:"discordWebhookUrl"`
	DiscordPrefixMessage string `json:"discordPrefixMessage"`
	DiscordDisableURL    bool   `json:"discordDisableUrl"`
	DiscordChannelType   string `json:"discordChannelType"`
	DiscordThreadID      string `json:"discordThreadId"`
	DiscordPostName      string `json:"discordPostName"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func maskWebhookURL(url string) string {
	if len(url) < 20 {
		return "***"
	}
	return url[:8] + "***" + url[len(url)-8:]
}

func ToNotification(n *model.Notification) *Notification {
	return &Notification{
		ID:                   n.ID.String(),
		HostID:               n.HostID.String(),
		Name:                 n.Name,
		Type:                 n.Type,
		Active:               *n.Active,
		LastNotifiedAt:       n.LastNotifiedAt.Time,
		DiscordUsername:      n.DiscordUsername,
		DiscordWebhookURL:    maskWebhookURL(n.DiscordWebhookURL),
		DiscordPrefixMessage: n.DiscordPrefixMessage,
		DiscordDisableURL:    n.DiscordDisableURL,
		DiscordChannelType:   n.DiscordChannelType,
		DiscordThreadID:      n.DiscordThreadID,
		DiscordPostName:      n.DiscordPostName,
		CreatedAt:            n.CreatedAt,
		UpdatedAt:            n.UpdatedAt,
	}
}

func ToNotifications(notifications []model.Notification) []Notification {
	results := make([]Notification, len(notifications))
	for i, n := range notifications {
		results[i] = *ToNotification(&n)
	}
	return results
}

type UpdateNotificationRequest struct {
	Name   string `json:"name" binding:"omitempty,max=100"`
	Type   string `json:"type" binding:"omitempty,oneof=discord"`
	Active bool   `json:"active" binding:"omitempty"`

	DiscordUsername      string `json:"discordUsername" binding:"omitempty"`
	DiscordWebhookURL    string `json:"discordWebhookUrl" binding:"omitempty,url"`
	DiscordPrefixMessage string `json:"discordPrefixMessage" binding:"omitempty"`
	DiscordDisableURL    bool   `json:"discordDisableUrl" binding:"omitempty"`
	DiscordChannelType   string `json:"discordChannelType" binding:"omitempty,oneof=postToThread createNewForumPost"`
	DiscordThreadID      string `json:"discordThreadId" binding:"omitempty"`
	DiscordPostName      string `json:"discordPostName" binding:"omitempty,max=100"`
}

func (r UpdateNotificationRequest) ToModel() *model.Notification {
	return &model.Notification{
		Name:                 r.Name,
		Type:                 model.NotificationType(r.Type),
		Active:               &r.Active,
		DiscordWebhookURL:    r.DiscordWebhookURL,
		DiscordPrefixMessage: r.DiscordPrefixMessage,
		DiscordUsername:      r.DiscordUsername,
		DiscordDisableURL:    r.DiscordDisableURL,
		DiscordChannelType:   r.DiscordChannelType,
		DiscordThreadID:      r.DiscordThreadID,
		DiscordPostName:      r.DiscordPostName,
	}
}
