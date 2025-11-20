package notification

import (
	"context"
	"log"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/notification/discord"
	"github.com/DarknessKiller/pingopher/internal/repository"
)

type NotificationService struct {
	repo Repository
}

func NewService(hostRepository *repository.BaseRepository[model.Host], historyRepo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: &newRepository{HostRepo: hostRepository, NotificationRepo: historyRepo}}
}

func (ns *NotificationService) CreateNotification(ctx context.Context, hostId string, notification *model.Notification) error {
	host, err := ns.repo.Host().GetByID(ctx, hostId)
	if err != nil {
		return err
	}

	notification.HostID = host.ID
	return ns.repo.Notification().Create(ctx, notification)
}

func (ns *NotificationService) GetNotificationsForHost(ctx context.Context, hostID string) ([]model.Notification, error) {
	return ns.repo.Notification().GetActiveNotificationsForHost(ctx, hostID)
}

func (ns *NotificationService) DeleteNotification(ctx context.Context, id string) error {
	return ns.repo.Notification().Delete(ctx, id)
}

func (ns *NotificationService) UpdateNotification(ctx context.Context, id string, notification *model.Notification) error {
	return ns.repo.Notification().Update(ctx, id, notification)
}

func (ns *NotificationService) SendNotification(host *model.Host, histories []*model.History) {
	ctx := context.Background()
	notifications, err := ns.repo.Notification().GetActiveNotificationsForHost(ctx, host.ID.String())
	if err != nil {
		log.Printf("[%s] failed to load notifications: %v", host.HostURL, err)
		return
	}

	for _, n := range notifications {
		if err := ns.sendOne(&n, host, histories); err != nil {
			log.Printf("[%s] failed to send notification (type=%s): %v", host.HostURL, n.Type, err)
		}

		if updateErr := ns.repo.Notification().UpdateLastNotifiedAt(ctx, n.ID.String()); updateErr != nil {
			log.Printf("[%s] failed to update LastNotifiedAt: %v", host.HostURL, updateErr)
		}
	}
}

func (ns *NotificationService) sendOne(n *model.Notification, host *model.Host, histories []*model.History) error {
	switch n.Type {
	case model.DiscordNotification:
		d := discord.DiscordNotification{
			WebhookURL:    n.DiscordWebhookURL,
			Username:      n.DiscordUsername,
			PrefixMessage: n.DiscordPrefixMessage,
			DisableURL:    n.DiscordDisableURL,
			ChannelType:   n.DiscordChannelType,
			ThreadID:      n.DiscordThreadID,
			PostName:      n.DiscordPostName,
		}
		return d.Send(host, histories)
	default:
		log.Printf("[%s] unknown notification type: %s", host.HostURL, n.Type)
		return nil
	}
}
