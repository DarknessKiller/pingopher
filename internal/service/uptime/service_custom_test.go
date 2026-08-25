package uptime_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/service/uptime"
	"github.com/segmentio/ksuid"
)

func newUptimeService() (*uptime.Service, *MockHostRepository, *MockHistoryRepository) {
	hostRepo := &MockHostRepository{}
	historyRepo := &MockHistoryRepository{}
	mockRedis := &MockRedis{}
	cfg := &config.Config{Env: "test"}
	svc := uptime.NewService(cfg, hostRepo, historyRepo, mockRedis)

	hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
		return nil, errors.New("not found")
	}
	historyRepo.CreatePingHistoryFunc = func(ctx context.Context, previousStatus model.HostStatus, host *model.Host, histories []*model.History) ([]*model.History, error) {
		return histories, nil
	}

	return svc, hostRepo, historyRepo
}

func TestService_PingHost_TCP(t *testing.T) {
	svc, hostRepo, _ := newUptimeService()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	port := uint16(ln.Addr().(*net.TCPAddr).Port)

	t.Run("TCP_Success", func(t *testing.T) {
		hostID := ksuid.New()
		hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
			return &model.Host{
				BaseModel: model.BaseModel{ID: hostID},
				HostURL:   "127.0.0.1",
				Port:      &port,
				Protocol:  "tcp",
				Status:    model.HostStatusUnknown,
			}, nil
		}

		_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if host.Status != model.HostStatusUp {
			t.Errorf("expected UP, got %v", host.Status)
		}
		if histories[0].StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", histories[0].StatusCode)
		}
	})

	t.Run("TCP_Failure", func(t *testing.T) {
		deadLn, _ := net.Listen("tcp", "127.0.0.1:0")
		deadPort := uint16(deadLn.Addr().(*net.TCPAddr).Port)
		deadLn.Close() // nothing listening

		hostID := ksuid.New()
		hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
			return &model.Host{
				BaseModel: model.BaseModel{ID: hostID},
				HostURL:   "127.0.0.1",
				Port:      &deadPort,
				Protocol:  "tcp",
				Status:    model.HostStatusUnknown,
			}, nil
		}

		_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if host.Status != model.HostStatusDown {
			t.Errorf("expected DOWN, got %v", host.Status)
		}
		if histories[0].StatusCode != 0 {
			t.Errorf("expected status code 0, got %d", histories[0].StatusCode)
		}
		if histories[0].ErrorMessage == "" {
			t.Error("expected an error message for a refused connection")
		}
	})
}

func TestService_PingHost_UDP(t *testing.T) {
	svc, hostRepo, _ := newUptimeService()

	// UDP echo server: anything read back counts as alive.
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			conn.WriteTo(buf[:n], addr)
		}
	}()

	port := uint16(conn.LocalAddr().(*net.UDPAddr).Port)
	hostID := ksuid.New()
	hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
		return &model.Host{
			BaseModel: model.BaseModel{ID: hostID},
			HostURL:   "127.0.0.1",
			Port:      &port,
			Protocol:  "udp",
			Status:    model.HostStatusUnknown,
		}, nil
	}

	_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if host.Status != model.HostStatusUp {
		t.Errorf("expected UP, got %v", host.Status)
	}
	if histories[0].StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", histories[0].StatusCode)
	}
}

func TestService_PingHost_ICMP(t *testing.T) {
	svc, hostRepo, _ := newUptimeService()

	// Resolve ICMP targets against the default resolver, but stub the actual
	// ping so no real ICMP socket is needed. 127.0.0.1 is a literal IP.
	hostID := ksuid.New()
	hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
		return &model.Host{
			BaseModel: model.BaseModel{ID: hostID},
			HostURL:   "127.0.0.1",
			Protocol:  "ping",
			Status:    model.HostStatusUnknown,
		}, nil
	}

	t.Run("ICMP_Success", func(t *testing.T) {
		uptime.SetPingFuncForTest(func(ctx context.Context, target string) (time.Duration, error) {
			if target != "127.0.0.1" {
				t.Errorf("expected target 127.0.0.1, got %s", target)
			}
			return 5 * time.Millisecond, nil
		})
		defer uptime.ResetPingFuncForTest()

		_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if host.Status != model.HostStatusUp {
			t.Errorf("expected UP, got %v", host.Status)
		}
		if histories[0].StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", histories[0].StatusCode)
		}
		if histories[0].Latency == 0 {
			t.Error("expected non-zero latency")
		}
	})

	t.Run("ICMP_Failure", func(t *testing.T) {
		uptime.SetPingFuncForTest(func(ctx context.Context, target string) (time.Duration, error) {
			return 0, errors.New("ping: no reply from 127.0.0.1")
		})
		defer uptime.ResetPingFuncForTest()

		_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if host.Status != model.HostStatusDown {
			t.Errorf("expected DOWN, got %v", host.Status)
		}
		if histories[0].StatusCode != 0 {
			t.Errorf("expected status code 0, got %d", histories[0].StatusCode)
		}
	})
}

func TestService_PingHost_IPv6_HTTP(t *testing.T) {
	svc, hostRepo, _ := newUptimeService()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	p, _ := strconv.ParseUint(u.Port(), 10, 16)
	port := uint16(p)

	hostID := ksuid.New()
	hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
		return &model.Host{
			BaseModel:           model.BaseModel{ID: hostID},
			HostURL:             u.Hostname(),
			Port:                &port,
			Protocol:            "http",
			Status:              model.HostStatusUnknown,
			AcceptedStatusCodes: []string{"200-299"},
		}, nil
	}

	_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if host.Status != model.HostStatusUp {
		t.Errorf("expected UP, got %v", host.Status)
	}
	if histories[0].StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", histories[0].StatusCode)
	}
}

func TestPingHostDeadline(t *testing.T) {
	svc, hostRepo, _ := newUptimeService()

	hostID := ksuid.New()
	hostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
		return &model.Host{
			BaseModel: model.BaseModel{ID: hostID},
			HostURL:   "192.0.2.1", // TEST-NET, guaranteed unroutable
			Protocol:  "tcp",
			Status:    model.HostStatusUnknown,
		}, nil
	}

	start := time.Now()
	_, _, _, err := svc.PingHost(context.Background(), hostID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("ping without caller deadline should self-timeout, took %v", elapsed)
	}
}
