package notification

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/cache"
	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/notification/discord"
	"github.com/DarknessKiller/pingopher/internal/repository"
)

type NotificationService struct {
	repository Repository
	cache      cache.Cache

	// cooldown caps alert sends per host and direction. Without it a flapping
	// or long-down host fires a webhook, and a D1 write for lastNotifiedAt, on
	// every single check.
	cooldown  time.Duration
	mu        sync.Mutex
	lastFired map[string]time.Time
}

func NewService(config *config.Config, hostRepository *repository.BaseRepository[model.Host], historyRepo repository.NotificationRepository, cacheClient cache.Cache) *NotificationService {
	return &NotificationService{
		repository: &newRepository{HostRepo: hostRepository, NotificationRepo: historyRepo},
		cache:      cacheClient,
		cooldown:   time.Duration(config.NotificationCooldown) * time.Second,
		lastFired:  make(map[string]time.Time),
	}
}

func (ns *NotificationService) CreateNotification(ctx context.Context, hostId string, notification *model.Notification) error {
	cacheKey := "pingopher_host:" + hostId
	var host *model.Host
	if err := ns.cache.Get(ctx, cacheKey, &host); err != nil {
		host, err = ns.repository.Host().GetByID(ctx, hostId)
		if err != nil {
			return err
		}

		_ = ns.cache.Set(ctx, cacheKey, host, 24*time.Hour)
	}

	notification.HostID = host.ID
	err := ns.repository.Notification().Create(ctx, notification)
	if err == nil {
		_ = ns.cache.Delete(ctx, "pingopher_notifications:"+hostId, "pingopher_active_notifications:"+hostId)
	}
	return err
}

func (ns *NotificationService) GetNotificationsForHost(ctx context.Context, hostId string) ([]model.Notification, error) {
	cacheKey := "pingopher_notifications:" + hostId
	var notifications []model.Notification
	if err := ns.cache.Get(ctx, cacheKey, &notifications); err == nil {
		return notifications, nil
	}

	notifications, err := ns.repository.Notification().GetNotificationsForHost(ctx, hostId)
	if err != nil {
		return nil, err
	}

	_ = ns.cache.Set(ctx, cacheKey, notifications, 24*time.Hour)
	return notifications, nil
}

func (ns *NotificationService) DeleteNotification(ctx context.Context, hostId, notificationId string) error {
	notification, err := ns.repository.Notification().GetByID(ctx, notificationId)
	if err != nil {
		return err
	}

	if notification.HostID.String() != hostId {
		return errors.New("notification not found")
	}

	err = ns.repository.Notification().Delete(ctx, notificationId)
	if err == nil {
		_ = ns.cache.Delete(ctx, "pingopher_notifications:"+hostId, "notification:"+notificationId)
	}
	return err
}

func (ns *NotificationService) UpdateNotification(ctx context.Context, hostId, notificationId string, notification *model.Notification) error {
	oldNotification, err := ns.repository.Notification().GetByID(ctx, notificationId)
	if err != nil {
		return err
	}

	if oldNotification.HostID.String() != hostId {
		return errors.New("notification not found")
	}

	err = ns.repository.Notification().Update(ctx, notificationId, notification)
	if err == nil {
		_ = ns.cache.Delete(ctx, "pingopher_notifications:"+hostId, "pingopher_active_notifications:"+hostId)
	}
	return err
}

// allowFire reports whether an alert for this host and status may go out now,
// and records the fire. Down and up keep separate buckets, so a throttled
// down alert never swallows the recovery alert.
func (ns *NotificationService) allowFire(hostID string, status model.HostStatus) bool {
	key := hostID + ":" + string(status)

	ns.mu.Lock()
	defer ns.mu.Unlock()

	if last, ok := ns.lastFired[key]; ok && time.Since(last) < ns.cooldown {
		return false
	}

	ns.lastFired[key] = time.Now()
	return true
}

func (ns *NotificationService) SendNotification(host *model.Host, histories []*model.History) {
	ctx := context.Background()
	if !ns.allowFire(host.ID.String(), host.Status) {
		log.Printf("[%s] notification suppressed: %s cooldown", host.HostURL, ns.cooldown)
		return
	}

	cacheKey := "pingopher_active_notifications:" + host.ID.String()
	var notifications []model.Notification
	if err := ns.cache.Get(ctx, cacheKey, &notifications); err != nil {
		notifications, err = ns.repository.Notification().GetActiveNotificationsForHost(ctx, host.ID.String())
		if err != nil {
			log.Printf("[%s] failed to load notifications: %v", host.HostURL, err)
			return
		}

		_ = ns.cache.Set(ctx, cacheKey, notifications, 24*time.Hour)
	}

	for _, n := range notifications {
		if err := ns.sendOne(&n, host, histories); err != nil {
			log.Printf("[%s] failed to send notification (type=%s): %v", host.HostURL, n.Type, err)
		}

		if updateErr := ns.repository.Notification().UpdateLastNotifiedAt(ctx, n.ID.String()); updateErr != nil {
			log.Printf("[%s] failed to update LastNotifiedAt: %v", host.HostURL, updateErr)
		}
	}
}

func (ns *NotificationService) DeleteNotifications(ctx context.Context, hostId string) error {
	err := ns.repository.Notification().DeleteNotifications(ctx, hostId)
	if err == nil {
		_ = ns.cache.Delete(ctx, "pingopher_notifications:"+hostId, "pingopher_active_notifications:"+hostId)
	}
	return err
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

func (ns *NotificationService) SendSystemErrorNotification(host *model.Host, title, errorMessage string) {
	ctx := context.Background()
	cacheKey := "pingopher_active_notifications:" + host.ID.String()
	var notifications []model.Notification
	if err := ns.cache.Get(ctx, cacheKey, &notifications); err != nil {
		notifications, err = ns.repository.Notification().GetActiveNotificationsForHost(ctx, host.ID.String())
		if err != nil {
			log.Printf("[%s] failed to load notifications for system error: %v", host.HostURL, err)
			return
		}

		_ = ns.cache.Set(ctx, cacheKey, notifications, 24*time.Hour)
	}

	for _, n := range notifications {
		if n.Type == model.DiscordNotification {
			if err := discord.SendSystemErrorWebhook(n.DiscordWebhookURL, title, errorMessage); err != nil {
				log.Printf("[%s] failed to send system error webhook to Discord: %v", host.HostURL, err)
			}
		}
	}
}
