package notification_test

import (
	"testing"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
)

func TestNotifyCooldown(t *testing.T) {
	ns := notification.NewService(&config.Config{}, nil, nil, nil)
	notification.SetCooldown(ns, 20*time.Millisecond)

	const host = "host-1"
	down := model.HostStatusDown

	if !notification.AllowFire(ns, host, down) {
		t.Fatal("first down alert must fire")
	}
	if notification.AllowFire(ns, host, down) {
		t.Fatal("repeat down alert inside the cooldown must be suppressed")
	}
	if !notification.AllowFire(ns, host, model.HostStatusUp) {
		t.Fatal("down cooldown must not swallow the recovery alert")
	}
	if !notification.AllowFire(ns, "host-2", down) {
		t.Fatal("one host's cooldown must not silence another host")
	}

	time.Sleep(25 * time.Millisecond)

	if !notification.AllowFire(ns, host, down) {
		t.Fatal("alert after the cooldown window must fire")
	}
}
