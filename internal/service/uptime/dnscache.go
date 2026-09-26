package uptime

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
)

// dnsCache reuses addresses until the server-supplied TTL expires; entries are keyed per resolver +
// hostname. Errors are never cached and TTL 0 records are not stored.
type dnsCache struct {
	mu      sync.Mutex
	entries map[dnsKey]dnsEntry
}

type dnsKey struct {
	dns  model.DNS
	host string
}

type dnsEntry struct {
	ips     []net.IPAddr
	expires time.Time
}

func newDNSCache() *dnsCache {
	return &dnsCache{entries: make(map[dnsKey]dnsEntry)}
}

// lookup serves cached addresses until their TTL expires; the whole answer set ages out on its
// shortest record TTL, matching how a normal resolver ages a response.
func (c *dnsCache) lookup(ctx context.Context, dns model.DNS, host string, resolve resolveFunc) ([]net.IPAddr, error) {
	key := dnsKey{dns: dns, host: host}

	c.mu.Lock()
	if entry, ok := c.entries[key]; ok && time.Now().Before(entry.expires) {
		c.mu.Unlock()
		return entry.ips, nil
	}
	c.mu.Unlock()

	records, err := resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}

	ips := make([]net.IPAddr, 0, len(records))
	ttl := records[0].TTL
	for _, record := range records {
		ips = append(ips, net.IPAddr{IP: record.IP})
		if record.TTL < ttl {
			ttl = record.TTL
		}
	}

	if ttl > 0 {
		c.mu.Lock()
		c.entries[key] = dnsEntry{ips: ips, expires: time.Now().Add(ttl)}
		c.mu.Unlock()
	}

	return ips, nil
}
