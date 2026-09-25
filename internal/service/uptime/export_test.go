package uptime

import (
	"context"
	"net"

	"github.com/DarknessKiller/pingopher/internal/model"
)

// Test seams. Tests for this package are black-box (package uptime_test), so
// the unexported resolver internals they cover are re-exported here. This file
// is compiled into the test binary only, never into the server.

type (
	DNSCache  = dnsCache
	DNSRecord = dnsRecord
)

var (
	NewDNSCache   = newDNSCache
	ParseResponse = parseResponse
	QueryServer   = queryServer
)

// Lookup is dnsCache.lookup: an unexported method cannot be called from another
// package, so it is wrapped rather than aliased.
func Lookup(cache *DNSCache, ctx context.Context, dns model.DNS, host string, resolve func(context.Context, string) ([]DNSRecord, error)) ([]net.IPAddr, error) {
	return cache.lookup(ctx, dns, host, resolve)
}
