package uptime

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
	probing "github.com/prometheus-community/pro-bing"
)

// pingICMP is the default ICMP prober. Tests swap it via SetPingFuncForTest.
var pingICMP = defaultPingICMP

// SetPingFuncForTest replaces the ICMP prober for the duration of a test.
func SetPingFuncForTest(fn func(context.Context, string) (time.Duration, error)) {
	pingICMP = fn
}

// ResetPingFuncForTest restores the real ICMP prober.
func ResetPingFuncForTest() {
	pingICMP = defaultPingICMP
}

// defaultPingICMP probes target with a single ICMP echo request. It prefers the
// privileged raw-socket path (root, or NET_RAW capability in Docker) and
// falls back to the unprivileged datagram socket that Linux allows for
// ping_group_range.
// runPinger executes one echo request on the given pinger and reports
// latency. A run that sent the request but got no reply is a timeout, not a
// socket error.
func runPinger(ctx context.Context, pinger *probing.Pinger) (time.Duration, error) {
	err := pinger.RunWithContext(ctx)
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, fmt.Errorf("ping: no reply from %s (timeout)", pinger.Addr())
	}
	return stats.AvgRtt, nil
}

// defaultPingICMP probes target with a single ICMP echo request. It prefers the
// privileged raw-socket path (root, or NET_RAW capability in Docker) and
// falls back to the unprivileged datagram socket that Linux allows for
// ping_group_range.
func defaultPingICMP(ctx context.Context, target string) (time.Duration, error) {
	newPinger := func() (*probing.Pinger, error) {
		p, err := probing.NewPinger(target)
		if err != nil {
			return nil, err
		}
		p.Count = 1
		p.Timeout = time.Second * 2
		return p, nil
	}

	pinger, err := newPinger()
	if err != nil {
		return 0, err
	}
	pinger.SetPrivileged(true)

	if latency, runErr := runPinger(ctx, pinger); runErr == nil {
		return latency, nil
	}

	// Privileged path unavailable (no raw socket) — retry unprivileged.
	up, err := newPinger()
	if err != nil {
		return 0, err
	}
	up.SetPrivileged(false)

	return runPinger(ctx, up)
}

// buildICMPHistory resolves the target (honoring a custom DNS resolver) and
// pings it. StatusCode 200 means a reply came back; 0 means it failed.
// ping is injectable so tests can avoid real sockets.
func buildICMPHistory(ctx context.Context, host *model.Host, dns model.DNS, ping func(context.Context, string) (time.Duration, error)) *model.History {
	target, err := resolveICMPTarget(ctx, host.HostURL, dns)
	if err != nil {
		return &model.History{
			StatusCode:   0,
			HostID:       host.ID,
			DNS:          dns,
			ErrorMessage: err.Error(),
		}
	}

	latency, err := ping(ctx, target)
	if err != nil {
		return &model.History{
			StatusCode:   0,
			HostID:       host.ID,
			DNS:          dns,
			ErrorMessage: err.Error(),
		}
	}

	return &model.History{
		StatusCode: 200,
		Latency:    latencyMilliseconds(latency),
		HostID:     host.ID,
		DNS:        dns,
	}
}

// resolveICMPTarget returns a literal IP, or resolves a hostname through the
// configured DNS resolver so multi-DNS monitoring applies to ping targets too.
func resolveICMPTarget(ctx context.Context, host string, dns model.DNS) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return host, nil
	}

	resolver := net.DefaultResolver
	if dns != (model.DNS{Name: "System DNS"}) {
		dnsAddr := dns.IP
		if dns.Port != 0 {
			dnsAddr = net.JoinHostPort(dns.IP, strconv.Itoa(int(dns.Port)))
		}
		resolver = &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 500 * time.Millisecond}
				return d.DialContext(ctx, dns.Protocol, dnsAddr)
			},
		}
	}

	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no IPs found for %s", host)
	}

	return ips[0].IP.String(), nil
}

// ipFamily returns a best-effort "4" or "6" for a host so pro-bing can pick
// the matching socket. Empty means "default" (first resolved address).
func ipFamily(host string) string {
	host = strings.Trim(host, "[]")
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return "4"
		}
		return "6"
	}
	return ""
}
