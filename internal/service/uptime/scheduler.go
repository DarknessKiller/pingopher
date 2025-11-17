package uptime

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron     *cron.Cron
	hostJobs map[string]cron.EntryID
	mu       sync.Mutex
	service  *Service

	failCounts map[string]int
}

func NewScheduler(service *Service) *Scheduler {
	return &Scheduler{
		cron:       cron.New(),
		hostJobs:   make(map[string]cron.EntryID),
		service:    service,
		failCounts: make(map[string]int),
	}
}

func (ps *Scheduler) Start() {
	ctx := context.Background()
	hosts, _ := ps.service.GetAllHosts(ctx)

	for _, host := range hosts {
		ps.scheduleHost(ctx, &host)
	}

	ps.cron.Start()
	log.Println("Uptime monitoring scheduler started")
}

func (ps *Scheduler) ScheduleHost(ctx *gin.Context, host *model.Host) {
	ps.scheduleHost(ctx, host)
}

func (ps *Scheduler) DeleteHost(ctx context.Context, hostID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if jobID, ok := ps.hostJobs[hostID]; ok {
		ps.cron.Remove(jobID)
		delete(ps.hostJobs, hostID)
	}
}

func (ps *Scheduler) Stop() {
	ctx := ps.cron.Stop()
	<-ctx.Done()
	log.Println("Uptime monitoring scheduler stopped")
}

func (ps *Scheduler) scheduleHost(ctx context.Context, host *model.Host) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if jobID, ok := ps.hostJobs[host.ID.String()]; ok {
		ps.cron.Remove(jobID)
		delete(ps.hostJobs, host.ID.String())
	}

	interval := host.PingInterval
	if interval == 0 {
		return
	}

	spec := fmt.Sprintf("@every %ds", interval)

	jobID, err := ps.cron.AddFunc(spec, func() {
		ps.runPingForHost(ctx, host.ID.String())
	})
	if err != nil {
		log.Printf("[%s] failed to schedule: %ds", host.HostURL, err)
		return
	}

	ps.hostJobs[host.ID.String()] = jobID
	log.Printf("[%s] scheduled every %ds", host.HostURL, interval)
}

func (ps *Scheduler) runPingForHost(ctx context.Context, hostID string) {
	host, err := ps.service.GetHostByID(ctx, hostID)
	if err != nil {
		log.Printf("[%s] fetch error: %v", hostID, err)
		return
	}

	log.Printf("[%s] pinging...", host.HostURL)
	histories, newStatus, err := ps.service.PingHost(ctx, hostID)
	if err != nil {
		log.Printf("[%s] ping failed: %v", host.HostURL, err)
		return
	}

	prevStatus := host.Status
	if newStatus == "failed" {
		ps.failCounts[hostID]++
	} else {
		ps.failCounts[hostID] = 0
	}

	if prevStatus != newStatus {
		log.Printf("[%s] status changed: %s → %s", host.HostURL, prevStatus, newStatus)
		util.SendNotificationWebhook(host, histories)
		return
	}

	if newStatus == "failed" && ps.failCounts[hostID] > 0 && ps.failCounts[hostID]%int(host.FailThreshold) == 0 {
		log.Printf("[%s] still failed after %d checks — resending notification", host.HostURL, ps.failCounts[hostID])
		util.SendNotificationWebhook(host, histories)
		return
	}
}
