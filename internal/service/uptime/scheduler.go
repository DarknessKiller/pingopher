package uptime

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/service/notification"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type hostState struct {
	jobID         cron.EntryID
	jobSpec       string
	downStartTime time.Time
	downCount     int
	inFlight      bool
}

type Scheduler struct {
	cron                *cron.Cron
	states              map[string]*hostState
	mu                  sync.Mutex
	service             *Service
	notificationService *notification.NotificationService
}

func NewScheduler(service *Service, notificationService *notification.NotificationService) *Scheduler {
	return &Scheduler{
		cron:                cron.New(),
		states:              make(map[string]*hostState),
		service:             service,
		notificationService: notificationService,
	}
}

func (ps *Scheduler) getStateLocked(hostID string) *hostState {
	if state, ok := ps.states[hostID]; ok {
		return state
	}
	state := &hostState{}
	ps.states[hostID] = state
	return state
}

func (ps *Scheduler) Start() {
	ctx := context.Background()

	var (
		hosts []model.Host
		err   error
	)

	for {
		hosts, err = ps.service.GetAllHosts(ctx)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal(err)
		} else if err != nil {
			continue
		}
		break
	}

	for i := range hosts {
		ps.scheduleHost(ctx, &hosts[i])
	}

	ps.cron.Start()
	log.Println("Uptime monitoring scheduler started with dynamic backoff")
}

func (ps *Scheduler) ScheduleHost(ctx *gin.Context, host *model.Host) {
	ps.scheduleHost(ctx, host)
}

func (ps *Scheduler) DeleteHost(ctx context.Context, hostID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if state, ok := ps.states[hostID]; ok && state.jobID != 0 {
		ps.cron.Remove(state.jobID)
	}
	delete(ps.states, hostID)
}

func (ps *Scheduler) Stop() {
	ctx := ps.cron.Stop()
	<-ctx.Done()
	log.Println("Uptime monitoring scheduler stopped")
}

func (ps *Scheduler) scheduleHost(ctx context.Context, host *model.Host) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	state := ps.getStateLocked(host.ID.String())

	if host.PingInterval == 0 {
		if state.jobID != 0 {
			ps.cron.Remove(state.jobID)
			state.jobID = 0
			state.jobSpec = ""
		}
		return
	}

	interval := ps.calculatePingInterval(host, state)
	spec := fmt.Sprintf("@every %ds", int(interval.Seconds()))

	if state.jobID != 0 {
		if state.jobSpec == spec {
			return
		}
		ps.cron.Remove(state.jobID)
	}

	jobID, err := ps.cron.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ps.runPingForHost(ctx, host.ID.String())
	})
	if err != nil {
		log.Printf("[%s] failed to schedule: %v", host.HostURL, err)
		return
	}

	state.jobID = jobID
	state.jobSpec = spec
	log.Printf("[%s] scheduled every %ds", host.HostURL, int(interval.Seconds()))
}

func (ps *Scheduler) calculatePingInterval(host *model.Host, state *hostState) time.Duration {
	normalInterval := time.Duration(host.PingInterval) * time.Second

	if host.Status != model.HostStatusDown {
		return normalInterval
	}

	if state.downStartTime.IsZero() {
		state.downStartTime = time.Now()
		return time.Duration(ps.service.config.MaxRetryInterval) * time.Second
	}

	downtime := time.Since(state.downStartTime)

	backoffFactor := float64(downtime) / float64(time.Duration(host.FailThreshold)*normalInterval)

	interval := time.Duration(ps.service.config.MaxRetryInterval) + time.Duration(float64(ps.service.config.MaxRetryInterval)*backoffFactor)

	if interval > normalInterval {
		interval = normalInterval
	}

	return interval
}

func (ps *Scheduler) runPingForHost(ctx context.Context, hostID string) {
	ps.mu.Lock()
	state := ps.getStateLocked(hostID)
	if state.inFlight {
		ps.mu.Unlock()
		return
	}
	state.inFlight = true
	ps.mu.Unlock()

	defer func() {
		ps.mu.Lock()
		if state, ok := ps.states[hostID]; ok {
			state.inFlight = false
		}
		ps.mu.Unlock()
	}()

	prevStatus, host, histories, err := ps.service.PingHost(ctx, hostID)
	if err != nil {
		if strings.Contains(err.Error(), "Exceeded maximum DB size") || strings.Contains(err.Error(), "7500") {
			if host != nil {
				ps.notificationService.SendSystemErrorNotification(host, "Database Error", err.Error())
			}
		}
		if !(os.IsTimeout(err) || errors.Is(err, gorm.ErrRecordNotFound)) {
			log.Printf("[%s] ping failed: %v", hostID, err)
		}
		return
	}

	log.Printf("[%s] ping result: %s", host.HostURL, host.Status)

	ps.mu.Lock()
	if host.Status == model.HostStatusDown {
		state.downCount++
	} else {
		state.downCount = 0
	}
	downCount := state.downCount
	ps.mu.Unlock()

	if prevStatus != host.Status {
		log.Printf("[%s] status changed: %s → %s", host.HostURL, prevStatus, host.Status)

		ps.mu.Lock()
		if host.Status == model.HostStatusUp {
			state.downStartTime = time.Time{}
		} else {
			state.downStartTime = time.Now()
		}
		ps.mu.Unlock()

		ps.notificationService.SendNotification(host, histories)

		ps.scheduleHost(ctx, host)
		return
	}

	if host.Status == model.HostStatusDown && downCount > 0 && downCount%int(host.FailThreshold) == 0 {
		log.Printf("[%s] still failed after %d checks — resending notification", host.HostURL, downCount)
		ps.notificationService.SendNotification(host, histories)
	}

	ps.scheduleHost(ctx, host)
}
