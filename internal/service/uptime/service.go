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

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	InvalidateByPrefix(ctx context.Context, prefix string) error
}

type Service struct {
	config     *config.Config
	repository Repository
	cache      Cache
}

func NewService(config *config.Config, hostRepository repository.HostRepository, historyRepository repository.HistoryRepository, cacheClient Cache) *Service {
	return &Service{config: config, repository: &newRepository{HostRepo: hostRepository, HistoryRepo: historyRepository}, cache: cacheClient}
}

func (s *Service) CreateHost(ctx context.Context, host *model.Host) error {
	err := s.repository.Host().Create(ctx, host)
	if err == nil {
		s.cache.Delete(ctx, "pingopher_hosts:all")
	}
	return err
}

func (s *Service) GetHostByID(ctx context.Context, hostID string) (*model.Host, error) {
	cacheKey := "pingopher_host:" + hostID
	var host *model.Host
	if err := s.cache.Get(ctx, cacheKey, &host); err == nil {
		return host, nil
	}

	host, err := s.repository.Host().GetByID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, host, 24*time.Hour)
	return host, err
}

func (s *Service) GetAllHosts(ctx context.Context) ([]model.Host, error) {
	cacheKey := "pingopher_hosts:all"
	var hosts []model.Host
	if err := s.cache.Get(ctx, cacheKey, &hosts); err == nil {
		return hosts, nil
	}

	hosts, err := s.repository.Host().GetAll(ctx)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, hosts, 24*time.Hour)
	}
	return hosts, err
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

	_ = s.cache.Delete(ctx, "pingopher_hosts:all", "pingopher_host:"+hostID)
	return host, nil
}

func (s *Service) DeleteHost(ctx context.Context, hostID string) error {
	_, err := s.repository.Host().GetByID(ctx, hostID)
	if err != nil {
		return err
	}

	err = s.repository.Host().Delete(ctx, hostID)
	if err == nil {
		_ = s.cache.Delete(ctx, "pingopher_hosts:all", "pingopher_host:"+hostID)
	}
	return err
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
	cacheKey := "pingopher_host:" + hostID
	if err := s.cache.Get(ctx, cacheKey, &host); err != nil {
		host, err = s.repository.Host().GetByID(ctx, hostID)
		if err != nil {
			return "", nil, nil, err
		}

		_ = s.cache.Set(ctx, cacheKey, host, 24*time.Hour)
	}

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)

	dnsList := []model.DNS{{Name: "System DNS"}}
	if len(host.DNS) > 0 {
		dnsList = host.DNS
	}

	pingTime := time.Now()

	histories = make([]*model.History, len(dnsList))

	for i, dns := range dnsList {
		wg.Add(1)
		go func(i int, dns model.DNS) {
			defer wg.Done()

			history := s.makeRequestAndBuildHistory(ctx, host, dns)
			history.PingDateTime = sql.NullTime{Time: pingTime, Valid: true}

			mutex.Lock()
			histories[i] = history
			mutex.Unlock()
		}(i, dns)
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

	if prevStatus != host.Status {
		_ = s.cache.Set(ctx, cacheKey, host, 24*time.Hour)
		_ = s.cache.Delete(ctx, "pingopher_hosts:all")
	}

	if histories, err = s.repository.History().CreatePingHistory(ctx, prevStatus, host, histories); err != nil {
		return prevStatus, host, histories, err
	}

	return prevStatus, host, histories, nil
}

func (s *Service) makeRequestAndBuildHistory(ctx context.Context, host *model.Host, dns model.DNS) *model.History {
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
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: 500 * time.Millisecond}
					return d.DialContext(ctx, dns.Protocol, dnsIP)
				},
			},
		}

		client = resty.NewWithDialer(dialer).SetTimeout(10 * time.Second)
		defer client.Close()
	}

	hostURL := strings.ToLower(host.Protocol) + "://" + host.HostURL
	if host.Port != nil && *host.Port != 0 {
		hostURL += ":" + strconv.Itoa(int(*host.Port))
	}

	if host.TLS.NoVerify != nil && *host.TLS.NoVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	var errorMsg string
	resp, err := client.R().WithContext(ctx).Get(hostURL)
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
