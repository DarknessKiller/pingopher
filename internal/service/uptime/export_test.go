package uptime

import (
	"context"
	"net"

	"github.com/DarknessKiller/pingopher/internal/model"
)

type (
	DNSCache  = dnsCache
	DNSRecord = dnsRecord
)

var (
	NewDNSCache   = newDNSCache
	ParseResponse = parseResponse
	QueryServer   = queryServer
)

func Lookup(cache *DNSCache, ctx context.Context, dns model.DNS, host string, resolve func(context.Context, string) ([]DNSRecord, error)) ([]net.IPAddr, error) {
	return cache.lookup(ctx, dns, host, resolve)
}
