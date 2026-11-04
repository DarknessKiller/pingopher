package uptime

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
)

// maxStaleAge stops a stale answer from serving forever when every refresh keeps failing.
const maxStaleAge = time.Hour

// dnsCache reuses addresses until the server-supplied TTL expires, keyed per resolver + hostname.
// Optimistic mode serves the expired answer while it refreshes; errors and TTL 0 answers are never stored.
type dnsCache struct {
	mu         sync.Mutex
	entries    map[dnsKey]dnsEntry
	refreshing map[dnsKey]bool
	optimistic bool
}

type dnsKey struct {
	dns  model.DNS
	host string
}

type dnsEntry struct {
	ips     []net.IPAddr
	expires time.Time
}

func newDNSCache(optimistic bool) *dnsCache {
	return &dnsCache{
		entries:    make(map[dnsKey]dnsEntry),
		refreshing: make(map[dnsKey]bool),
		optimistic: optimistic,
	}
}

// lookup serves cached addresses until their TTL expires; the whole answer set ages out on its
// shortest record TTL, matching how a normal resolver ages a response.
func (c *dnsCache) lookup(ctx context.Context, dns model.DNS, host string, resolve resolveFunc) ([]net.IPAddr, error) {
	key := dnsKey{dns: dns, host: host}
	c.mu.Lock()
	entry, ok := c.entries[key]
	if ok && time.Now().Before(entry.expires) {
		c.mu.Unlock()
		return entry.ips, nil
	}
	if ok && c.optimistic && time.Since(entry.expires) < maxStaleAge {
		if !c.refreshing[key] {
			c.refreshing[key] = true
			go c.refresh(ctx, key, host, resolve)
		}
		c.mu.Unlock()
		return entry.ips, nil
	}
	c.mu.Unlock()

	records, err := resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	return c.store(key, host, records)
}

// store caches a fresh answer and returns its addresses; TTL 0 answers are not stored.
func (c *dnsCache) store(key dnsKey, host string, records []dnsRecord) ([]net.IPAddr, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}
	ips := make([]net.IPAddr, 0, len(records))
	ttl := records[0].TTL
	for _, record := range records {
		ips = append(ips, net.IPAddr{IP: record.IP})
		ttl = min(ttl, record.TTL)
	}
	if ttl > 0 {
		c.mu.Lock()
		c.entries[key] = dnsEntry{ips: ips, expires: time.Now().Add(ttl)}
		c.mu.Unlock()
	}
	return ips, nil
}

// refresh renews an expired entry behind a stale answer; a failed lookup keeps the stale entry.
// shortcut: a TTL 0 answer leaves the entry expired and re-queries per lookup; upgrade when that costs.
func (c *dnsCache) refresh(ctx context.Context, key dnsKey, host string, resolve resolveFunc) {
	defer func() {
		c.mu.Lock()
		delete(c.refreshing, key)
		c.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dnsQueryTimeout)
	defer cancel()
	records, err := resolve(ctx, host)
	if err != nil {
		log.Printf("DNS refresh for %s failed, serving the stale answer: %v", host, err)
		return
	}
	if _, err := c.store(key, host, records); err != nil {
		log.Printf("DNS refresh for %s returned no addresses, serving the stale answer: %v", host, err)
		return
	}
}
