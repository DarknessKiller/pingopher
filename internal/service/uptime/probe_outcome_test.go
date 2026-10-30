package uptime_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/segmentio/ksuid"
)

func TestPingHost_NoProbeSent(t *testing.T) {
	hostID := ksuid.New()
	mockHost := &model.Host{
		BaseModel: model.BaseModel{ID: hostID},
		HostURL:   "192.0.2.1", // TEST-NET-1: nothing here answers, but this round never gets to ask
		Protocol:  "ping",
		Status:    model.HostStatusUp,
	}
	hostRepo := &MockHostRepository{
		GetByIDFunc: func(context.Context, string) (*model.Host, error) { return mockHost, nil },
	}
	historyRepo := &MockHistoryRepository{
		CreatePingHistoryFunc: func(context.Context, model.HostStatus, *model.Host, []*model.History) ([]*model.History, error) {
			t.Error("history written for a round that sent no probe")
			return nil, nil
		},
	}
	svc := uptime.NewService(&config.Config{Env: "test"}, hostRepo, historyRepo, &MockRedis{})

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	if _, _, _, err := svc.PingHost(ctx, hostID.String()); err == nil {
		t.Fatal("expected an error when the round has no budget left")
	}
	if mockHost.Status != model.HostStatusUp {
		t.Errorf("status changed to %v without a probe", mockHost.Status)
	}
}

func TestPingHost_TimedOutProbeIsPersisted(t *testing.T) {
	// A bound but silent UDP socket: the probe is sent, the reply never comes.
	socket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()
	port := uint16(socket.LocalAddr().(*net.UDPAddr).Port)

	hostID := ksuid.New()
	mockHost := &model.Host{
		BaseModel: model.BaseModel{ID: hostID},
		HostURL:   "127.0.0.1",
		Port:      &port,
		Protocol:  "udp",
		Status:    model.HostStatusUp,
	}
	hostRepo := &MockHostRepository{
		GetByIDFunc: func(context.Context, string) (*model.Host, error) { return mockHost, nil },
	}
	historyRepo := &MockHistoryRepository{
		CreatePingHistoryFunc: func(ctx context.Context, _ model.HostStatus, _ *model.Host, histories []*model.History) ([]*model.History, error) {
			if err := ctx.Err(); err != nil {
				t.Errorf("history written with the probe's expired context: %v", err)
			}
			return histories, nil
		},
	}
	svc := uptime.NewService(&config.Config{Env: "test"}, hostRepo, historyRepo, &MockRedis{})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	_, host, histories, err := svc.PingHost(ctx, hostID.String())
	if err != nil {
		t.Fatalf("a sent probe that timed out must still persist: %v", err)
	}
	if host.Status != model.HostStatusDown {
		t.Errorf("expected down, got %v", host.Status)
	}
	if len(histories) != 1 {
		t.Fatalf("expected the timed-out round to be recorded, got %d histories", len(histories))
	}
}
