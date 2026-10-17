package uptime

import (
	"context"
	"net"

	"github.com/DarknessKiller/pingopher/internal/model"
	"resty.dev/v3"
)

type (
	DNSCache  = dnsCache
	DNSRecord = dnsRecord
)

var (
	NewDNSCache   = newDNSCache
	ParseResponse = parseResponse
	QueryServer   = queryServer
	// ClientForHTTP and InitHTTPClients expose the shared-client pool to tests without an export
	// surface on production code.
	ClientForHTTP   = (*Service).httpClient
	InitHTTPClients = func(s *Service) { s.httpClients = make(map[httpClientKey]*resty.Client) }
)

func Lookup(cache *DNSCache, ctx context.Context, dns model.DNS, host string, resolve func(context.Context, string) ([]DNSRecord, error)) ([]net.IPAddr, error) {
	return cache.lookup(ctx, dns, host, resolve)
}
