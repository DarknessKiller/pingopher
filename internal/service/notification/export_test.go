package notification

import (
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
)

// Test seams. Tests for this package are black-box (package notification_test),
// so the unexported cooldown gate is exposed here. This file is compiled into
// the test binary only, never into the server.

// AllowFire is NotificationService.allowFire: an unexported method cannot be
// called from another package, so it is wrapped rather than aliased.
func AllowFire(ns *NotificationService, hostID string, status model.HostStatus) bool {
	return ns.allowFire(hostID, status)
}

func SetCooldown(ns *NotificationService, d time.Duration) {
	ns.cooldown = d
}
