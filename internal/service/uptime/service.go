package uptime

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/cache"
	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/repository"
	"github.com/DarknessKiller/pingopher/internal/util"
	"resty.dev/v3"
)

// pingTimeout caps every monitor check when the caller hasn't provided a
// tighter deadline. Previously the HTTP handler's on-demand ping had no
// timeout at all and could hang on a stuck host.
const pingTimeout = 10 * time.Second

type Service struct {
	config     *config.Config
	repository Repository
	cache      cache.Cache
}

func NewService(config *config.Config, hostRepository repository.HostRepository, historyRepository repository.HistoryRepository, cacheClient cache.Cache) *Service {
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
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pingTimeout)
		defer cancel()
	}

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
		down, err := historyIsDown(host, history)
		if err != nil {
			return "", nil, nil, err
		}
		if down {
			anyDown = true
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
	switch strings.ToLower(host.Protocol) {
	case "tcp":
		return s.pingTCP(ctx, host, dns)
	case "udp":
		return s.pingUDP(ctx, host, dns)
	case "ping":
		return buildICMPHistory(ctx, host, dns, pingICMP)
	default:
		return s.pingHTTP(ctx, host, dns)
	}
}

// historyIsDown reports whether a single check failed for the host's protocol.
// HTTP/S match against the accepted status codes; TCP/UDP/ping are binary
// reachability checks where StatusCode 0 means the probe failed.
func historyIsDown(host *model.Host, history *model.History) (bool, error) {
	switch strings.ToLower(host.Protocol) {
	case "http", "https":
		ok, err := util.CheckStatusCode(history.StatusCode, host.AcceptedStatusCodes)
		return !ok, err
	default:
		return history.StatusCode == 0, nil
	}
}

func (s *Service) pingHTTP(ctx context.Context, host *model.Host, dns model.DNS) *model.History {
	client := resty.New()

	if dns != (model.DNS{Name: "System DNS"}) {
		client = resty.NewWithDialer(s.getDialer(dns)).SetTimeout(pingTimeout)
	}

	userAgent := "Pingopher/Alpha (https://github.com/DarknessKiller/pingopher)"
	client.SetHeader("User-Agent", userAgent).
		SetDebug(s.config.Env != "production")

	defer client.Close()

	hostURL := strings.ToLower(host.Protocol) + "://" + host.HostURL
	if host.Port != nil && *host.Port != 0 {
		hostURL = strings.ToLower(host.Protocol) + "://" + net.JoinHostPort(host.HostURL, strconv.Itoa(int(*host.Port)))
	}

	if host.TLS.NoVerify != nil && *host.TLS.NoVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	var errorMsg string
	resp, err := client.R().WithContext(ctx).Get(hostURL)
	switch err.(type) {
	case nil:
		if resp != nil && resp.CascadeError != nil {
			errorMsg = resp.CascadeError.Error()
		}
	default:
		errorMsg = err.Error()
	}

	statusCode := 0
	latency := time.Duration(0)
	if resp != nil {
		statusCode = resp.StatusCode()
		latency = resp.Duration()
	}

	return &model.History{
		StatusCode:   uint16(statusCode),
		Latency:      latencyMilliseconds(latency),
		HostID:       host.ID,
		DNS:          dns,
		ErrorMessage: errorMsg,
	}
}

func (s *Service) pingTCP(ctx context.Context, host *model.Host, dns model.DNS) *model.History {
	target := targetAddr(host)

	start := time.Now()
	conn, err := s.getDialer(dns).DialContext(ctx, "tcp", target)
	latency := time.Since(start)

	statusCode := uint16(200)
	errorMsg := ""
	if err != nil {
		statusCode = 0
		errorMsg = err.Error()
	} else {
		conn.Close()
	}

	return &model.History{
		StatusCode:   statusCode,
		Latency:      latencyMilliseconds(latency),
		HostID:       host.ID,
		DNS:          dns,
		ErrorMessage: errorMsg,
	}
}

// pingUDP sends a probe and waits for a reply. UDP is connectionless, so dial
// alone can't prove reachability. For port 53 the probe is a DNS query; other
// ports get raw bytes and any datagram back counts as alive.
// ponytail: quiet UDP services that never reply are flagged down even if alive;
// switch to a protocol-aware probe if false downs matter.
func (s *Service) pingUDP(ctx context.Context, host *model.Host, dns model.DNS) *model.History {
	target := targetAddr(host)

	start := time.Now()
	conn, err := s.getDialer(dns).DialContext(ctx, "udp", target)
	if err != nil {
		return &model.History{
			StatusCode:   0,
			Latency:      latencyMilliseconds(time.Since(start)),
			HostID:       host.ID,
			DNS:          dns,
			ErrorMessage: err.Error(),
		}
	}
	defer conn.Close()

	probe := []byte("pingopher-probe")
	if host.Port != nil && *host.Port == 53 {
		probe = buildDNSQuery(host.HostURL, 255)
	}

	if _, err := conn.Write(probe); err != nil {
		return &model.History{
			StatusCode:   0,
			Latency:      latencyMilliseconds(time.Since(start)),
			HostID:       host.ID,
			DNS:          dns,
			ErrorMessage: err.Error(),
		}
	}

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetReadDeadline(deadline)
	} else {
		conn.SetReadDeadline(time.Now().Add(pingTimeout))
	}

	buf := make([]byte, 1500)
	if _, err := conn.Read(buf); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			err = errors.New("UDP read timeout: no response received")
		}
		return &model.History{
			StatusCode:   0,
			Latency:      latencyMilliseconds(time.Since(start)),
			HostID:       host.ID,
			DNS:          dns,
			ErrorMessage: err.Error(),
		}
	}

	return &model.History{
		StatusCode: 200,
		Latency:    latencyMilliseconds(time.Since(start)),
		HostID:     host.ID,
		DNS:        dns,
	}
}

func targetAddr(host *model.Host) string {
	port := uint16(0)
	if host.Port != nil {
		port = *host.Port
	}
	return net.JoinHostPort(host.HostURL, strconv.Itoa(int(port)))
}

func (s *Service) getResolver(dns model.DNS) *net.Resolver {
	if dns == (model.DNS{Name: "System DNS"}) {
		return nil
	}

	dnsAddr := dns.IP
	if dns.Port != 0 {
		dnsAddr = net.JoinHostPort(dns.IP, strconv.Itoa(int(dns.Port)))
	}

	return &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 500 * time.Millisecond}
			return d.DialContext(ctx, dns.Protocol, dnsAddr)
		},
	}
}

func (s *Service) getDialer(dns model.DNS) *net.Dialer {
	return &net.Dialer{
		Timeout:  pingTimeout,
		Resolver: s.getResolver(dns),
	}
}

// latencyMilliseconds rounds a duration up to the nearest whole millisecond.
func latencyMilliseconds(d time.Duration) uint16 {
	if d <= 0 {
		return 0
	}
	return uint16((d + time.Millisecond - 1) / time.Millisecond)
}

// buildDNSQuery builds a minimal DNS query for the given domain and record
// type. Common types: 1=A, 28=AAAA, 255=ANY.
func buildDNSQuery(domain string, qtype uint16) []byte {
	// Header: Transaction ID, Flags (standard query), 1 question, 0 answers
	header := []byte{0xAA, 0xBB, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// Strip trailing dot (FQDN notation) to avoid empty label on split.
	domain = strings.TrimSuffix(domain, ".")

	var body []byte
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 {
			continue
		}
		body = append(body, byte(len(label)))
		body = append(body, label...)
	}
	body = append(body, 0) // root label

	query := make([]byte, 0, len(header)+len(body)+4)
	query = append(query, header...)
	query = append(query, body...)
	query = append(query,
		byte(qtype>>8), byte(qtype), // QTYPE
		0x00, 0x01, // QCLASS = IN
	)
	return query
}
