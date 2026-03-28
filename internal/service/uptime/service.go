package uptime

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/repository"
	"github.com/DarknessKiller/pingopher/internal/util"
	"resty.dev/v3"
)

type Service struct {
	config     *config.Config
	repository Repository
}

func NewService(config *config.Config, hostRepository repository.HostRepository, historyRepository repository.HistoryRepository) *Service {
	return &Service{config: config, repository: &newRepository{HostRepo: hostRepository, HistoryRepo: historyRepository}}
}

func (s *Service) CreateHost(ctx context.Context, host *model.Host) error {
	return s.repository.Host().Create(ctx, host)
}

func (s *Service) GetHostByID(ctx context.Context, hostID string) (*model.Host, error) {
	return s.repository.Host().GetByID(ctx, hostID)
}

func (s *Service) GetAllHosts(ctx context.Context) ([]model.Host, error) {
	return s.repository.Host().GetAll(ctx)
}

func (s *Service) UpdateHost(ctx context.Context, hostID string, host *model.Host) (*model.Host, error) {
	err := s.repository.Host().Update(ctx, hostID, host)
	if err != nil {
		return nil, err
	}

	host, err = s.repository.Host().GetByID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (s *Service) DeleteHost(ctx context.Context, hostID string) error {
	_, err := s.repository.Host().GetByID(ctx, hostID)
	if err != nil {
		return err
	}

	return s.repository.Host().Delete(ctx, hostID)
}

func (s *Service) GetHistoryByHostID(ctx context.Context, hostID, startAt, endAt string) ([]*model.History, error) {
	startTime, err := time.Parse(time.RFC3339, startAt)
	if err != nil {
		return nil, err
	}

	endTime, err := time.Parse(time.RFC3339, endAt)
	if err != nil {
		return nil, err
	}

	startTime = startTime.In(time.Local)
	endTime = endTime.In(time.Local)

	return s.repository.History().GetHistoryByHostID(ctx, hostID, startTime, endTime)
}

func (s *Service) PingHost(ctx context.Context, hostID string) (prevStatus model.HostStatus, host *model.Host, histories []*model.History, err error) {
	host, err = s.repository.Host().GetByID(ctx, hostID)
	if err != nil {
		return "", nil, nil, err
	}

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)

	dnsList := host.DNS
	if len(dnsList) == 0 {
		dnsList = []model.DNS{{Name: "System DNS"}}
	}

	pingTime := time.Now()

	histories = make([]*model.History, len(dnsList))

	for i, dns := range dnsList {
		wg.Add(1)
		go func() {
			defer wg.Done()

			history := s.makeRequestAndBuildHistory(host, dns)
			history.PingDateTime = sql.NullTime{Time: pingTime, Valid: true}

			mutex.Lock()
			histories[i] = history
			mutex.Unlock()
		}()
	}

	wg.Wait()

	anyDown := false
	for _, history := range histories {
		if ok, err := util.CheckStatusCode(history.StatusCode, host.AcceptedStatusCodes); !ok {
			anyDown = true
		} else if err != nil {
			return "", nil, nil, err
		}
	}

	prevStatus = host.Status
	if anyDown {
		host.Status = model.HostStatusDown
	} else {
		host.Status = model.HostStatusUp
	}

	if histories, err = s.repository.History().CreatePingHistory(ctx, host, histories); err != nil {
		return prevStatus, host, histories, err
	}

	return prevStatus, host, histories, nil
}

func (s *Service) makeRequestAndBuildHistory(host *model.Host, dns model.DNS) *model.History {
	userAgent := "Pingopher/Alpha (https://github.com/DarknessKiller/pingopher)"
	client := resty.New().
		SetHeader("User-Agent", userAgent).
		SetDebug(s.config.Env != "production")

	defer client.Close()

	if dns != (model.DNS{Name: "System DNS"}) {
		dnsIP := dns.IP
		if dns.Port != 0 {
			dnsIP += ":" + strconv.Itoa(int(dns.Port))
		}

		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: 10 * time.Second}
					return d.DialContext(ctx, dns.Protocol, dnsIP)
				},
			},
		}

		client = resty.NewWithDialer(dialer).SetTimeout(10 * time.Second)
		defer client.Close()
	}

	hostURL := strings.ToLower(host.Protocol) + "://" + host.HostURL
	if *host.Port != 0 {
		hostURL += ":" + strconv.Itoa(int(*host.Port))
	}

	if host.TLS.NoVerify != nil && *host.TLS.NoVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	var errorMsg string
	resp, err := client.R().Get(hostURL)
	switch err.(type) {
	case nil:
		if resp.Err != nil {
			errorMsg = resp.Err.Error()
		}
	default:
		errorMsg = err.Error()
	}

	return &model.History{
		StatusCode:   uint16(resp.StatusCode()),
		Latency:      uint16(resp.Duration().Milliseconds()),
		HostID:       host.ID,
		DNS:          dns,
		ErrorMessage: errorMsg,
	}
}
