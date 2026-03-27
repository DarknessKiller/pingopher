package uptime_test

import (
	"context"
	"errors"
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

// MockHostRepository
type MockHostRepository struct {
	CreateFunc  func(ctx context.Context, host *model.Host) error
	GetByIDFunc func(ctx context.Context, id string) (*model.Host, error)
	GetAllFunc  func(ctx context.Context) ([]model.Host, error)
	UpdateFunc  func(ctx context.Context, id string, host *model.Host) error
	DeleteFunc  func(ctx context.Context, id string) error
}

func (m *MockHostRepository) Create(ctx context.Context, host *model.Host) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, host)
	}
	return nil
}

func (m *MockHostRepository) GetByID(ctx context.Context, id string) (*model.Host, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("mock GetByIDFunc not implemented")
}

func (m *MockHostRepository) GetAll(ctx context.Context) ([]model.Host, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, errors.New("mock GetAllFunc not implemented")
}

func (m *MockHostRepository) Update(ctx context.Context, id string, host *model.Host) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, host)
	}
	return errors.New("mock UpdateFunc not implemented")
}

func (m *MockHostRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("mock DeleteFunc not implemented")
}

// MockHistoryRepository
type MockHistoryRepository struct {
	CreateFunc             func(ctx context.Context, history *model.History) error
	GetByIDFunc            func(ctx context.Context, id string) (*model.History, error)
	GetAllFunc             func(ctx context.Context) ([]model.History, error)
	UpdateFunc             func(ctx context.Context, id string, history *model.History) error
	DeleteFunc             func(ctx context.Context, id string) error
	GetHistoryByIDFunc     func(ctx context.Context, historyId string) (*model.History, error)
	CreatePingHistoryFunc  func(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error)
	GetHistoryByHostIDFunc func(ctx context.Context, hostID string, startAt, endAt time.Time, status *[]string) ([]*model.History, error)
}

func (m *MockHistoryRepository) Create(ctx context.Context, history *model.History) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, history)
	}
	return nil
}

func (m *MockHistoryRepository) GetByID(ctx context.Context, id string) (*model.History, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("mock GetByIDFunc not implemented")
}

func (m *MockHistoryRepository) GetAll(ctx context.Context) ([]model.History, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, errors.New("mock GetAllFunc not implemented")
}

func (m *MockHistoryRepository) Update(ctx context.Context, id string, history *model.History) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, history)
	}
	return errors.New("mock UpdateFunc not implemented")
}

func (m *MockHistoryRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("mock DeleteFunc not implemented")
}

func (m *MockHistoryRepository) GetHistoryByID(ctx context.Context, historyId string) (*model.History, error) {
	if m.GetHistoryByIDFunc != nil {
		return m.GetHistoryByIDFunc(ctx, historyId)
	}
	return nil, errors.New("mock GetHistoryByIDFunc not implemented")
}

func (m *MockHistoryRepository) CreatePingHistory(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error) {
	if m.CreatePingHistoryFunc != nil {
		return m.CreatePingHistoryFunc(ctx, host, histories)
	}
	return nil, errors.New("mock CreatePingHistoryFunc not implemented")
}

func (m *MockHistoryRepository) GetHistoryByHostID(ctx context.Context, hostID string, startAt, endAt time.Time, status *[]string) ([]*model.History, error) {
	if m.GetHistoryByHostIDFunc != nil {
		return m.GetHistoryByHostIDFunc(ctx, hostID, startAt, endAt, status)
	}
	return nil, errors.New("mock GetHistoryByHostIDFunc not implemented")
}

func TestService_CreateHost(t *testing.T) {
	mockHostRepo := &MockHostRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	cfg := &config.Config{}
	svc := uptime.NewService(cfg, mockHostRepo, mockHistoryRepo)

	t.Run("Success", func(t *testing.T) {
		mockHostRepo.CreateFunc = func(ctx context.Context, host *model.Host) error {
			return nil
		}

		err := svc.CreateHost(context.Background(), &model.Host{Name: "Test Host"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockHostRepo.CreateFunc = func(ctx context.Context, host *model.Host) error {
			return expectedErr
		}

		err := svc.CreateHost(context.Background(), &model.Host{Name: "Test Host"})
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestService_PingHost(t *testing.T) {
	mockHostRepo := &MockHostRepository{}
	mockHistoryRepo := &MockHistoryRepository{}
	cfg := &config.Config{Env: "test"}
	svc := uptime.NewService(cfg, mockHostRepo, mockHistoryRepo)

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		p, _ := strconv.ParseUint(u.Port(), 10, 16)
		port := uint16(p)

		hostID := ksuid.New()
		mockHost := &model.Host{
			BaseModel:           model.BaseModel{ID: hostID},
			HostURL:             u.Hostname(),
			Port:                &port,
			Protocol:            "http",
			Status:              model.HostStatusUnknown,
			AcceptedStatusCodes: []string{"200-299"},
		}

		mockHostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
			if id == hostID.String() {
				return mockHost, nil
			}
			return nil, errors.New("not found")
		}

		mockHistoryRepo.CreatePingHistoryFunc = func(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error) {
			return histories, nil
		}

		prevStatus, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if prevStatus != model.HostStatusUnknown {
			t.Errorf("expected prevStatus %v, got %v", model.HostStatusUnknown, prevStatus)
		}
		if host.Status != model.HostStatusUp {
			t.Errorf("expected host status %v, got %v", model.HostStatusUp, host.Status)
		}
		if len(histories) == 0 {
			t.Error("expected histories, got empty")
		}
		if histories[0].StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", histories[0].StatusCode)
		}
	})

	t.Run("Failure_Down", func(t *testing.T) {
		// Create a server that returns 500
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		p, _ := strconv.ParseUint(u.Port(), 10, 16)
		port := uint16(p)

		hostID := ksuid.New()
		mockHost := &model.Host{
			BaseModel:           model.BaseModel{ID: hostID},
			HostURL:             u.Hostname(),
			Port:                &port,
			Protocol:            "http",
			Status:              model.HostStatusUnknown,
			AcceptedStatusCodes: []string{"200-299"},
		}

		mockHostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
			if id == hostID.String() {
				return mockHost, nil
			}
			return nil, errors.New("not found")
		}

		mockHistoryRepo.CreatePingHistoryFunc = func(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error) {
			return histories, nil
		}

		prevStatus, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if prevStatus != model.HostStatusUnknown {
			t.Errorf("expected prevStatus %v, got %v", model.HostStatusUnknown, prevStatus)
		}
		if host.Status != model.HostStatusDown {
			t.Errorf("expected host status %v, got %v", model.HostStatusDown, host.Status)
		}
		if len(histories) == 0 {
			t.Error("expected histories, got empty")
		}
		if histories[0].StatusCode != 500 {
			t.Errorf("expected status code 500, got %d", histories[0].StatusCode)
		}
	})

	t.Run("Failure_DBSizeExceeded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		p, _ := strconv.ParseUint(u.Port(), 10, 16)
		port := uint16(p)

		hostID := ksuid.New()
		mockHost := &model.Host{
			BaseModel:           model.BaseModel{ID: hostID},
			HostURL:             u.Hostname(),
			Port:                &port,
			Protocol:            "http",
			Status:              model.HostStatusUnknown,
			AcceptedStatusCodes: []string{"200-299"},
		}

		mockHostRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Host, error) {
			if id == hostID.String() {
				return mockHost, nil
			}
			return nil, errors.New("not found")
		}

		expectedErr := errors.New("Exceeded maximum DB size")
		mockHistoryRepo.CreatePingHistoryFunc = func(ctx context.Context, host *model.Host, histories []*model.History) ([]*model.History, error) {
			return nil, expectedErr
		}

		_, host, histories, err := svc.PingHost(context.Background(), hostID.String())
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if host == nil {
			t.Errorf("expected host to be returned even on db size error")
		}
		if len(histories) != 0 {
			t.Errorf("expected histories to be returned even on db size error")
		}
	})
}
