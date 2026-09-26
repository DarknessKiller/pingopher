package notification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/repository"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
	"github.com/segmentio/ksuid"
)

type stubCache struct{}

func (stubCache) Get(context.Context, string, interface{}) error { return errors.New("cache miss") }

func (stubCache) Set(context.Context, string, interface{}, time.Duration) error { return nil }

func (stubCache) Delete(context.Context, ...string) error { return nil }

func (stubCache) InvalidateByPrefix(context.Context, string) error { return nil }

type stubNotificationRepository struct {
	repository.NotificationRepository
	activeCalls int
}

func (r *stubNotificationRepository) GetActiveNotificationsForHost(context.Context, string) ([]model.Notification, error) {
	r.activeCalls++
	return nil, nil
}

func TestNotifyCooldown(t *testing.T) {
	const cooldownSeconds = 1

	repo := &stubNotificationRepository{}
	ns := notification.NewService(&config.Config{NotificationCooldown: cooldownSeconds}, nil, repo, stubCache{})

	host := func(id ksuid.KSUID, status model.HostStatus) *model.Host {
		return &model.Host{BaseModel: model.BaseModel{ID: id}, HostURL: "example.test", Status: status}
	}

	down := host(ksuid.New(), model.HostStatusDown)

	ns.SendNotification(down, nil)
	if repo.activeCalls != 1 {
		t.Fatal("first down alert must be sent")
	}

	ns.SendNotification(down, nil)
	if repo.activeCalls != 1 {
		t.Fatal("repeat down alert inside the cooldown must be suppressed")
	}

	ns.SendNotification(host(down.ID, model.HostStatusUp), nil)
	if repo.activeCalls != 2 {
		t.Fatal("down cooldown must not swallow the recovery alert")
	}

	ns.SendNotification(host(ksuid.New(), model.HostStatusDown), nil)
	if repo.activeCalls != 3 {
		t.Fatal("one host's cooldown must not silence another host")
	}

	time.Sleep(cooldownSeconds*time.Second + 100*time.Millisecond)

	ns.SendNotification(down, nil)
	if repo.activeCalls != 4 {
		t.Fatal("alert after the cooldown window must be sent")
	}
}
