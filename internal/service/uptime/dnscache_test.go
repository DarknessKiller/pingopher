package uptime_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/service/uptime"
)

func TestDNSCacheReusesAnswerWithinTTL(t *testing.T) {
	cache := uptime.NewDNSCache()

	var calls int32
	resolve := func(ctx context.Context, host string) ([]uptime.DNSRecord, error) {
		atomic.AddInt32(&calls, 1)
		return []uptime.DNSRecord{{IP: net.ParseIP("192.0.2.1"), TTL: time.Hour}}, nil
	}

	// Same resolver + host repeatedly: exactly one lookup.
	for i := 0; i < 3; i++ {
		if _, err := uptime.Lookup(cache, context.Background(), model.DNS{Name: "System DNS"}, "example.com", resolve); err != nil {
			t.Fatal(err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 resolver call within TTL, got %d", got)
	}

	// A different resolver must not share the entry.
	if _, err := uptime.Lookup(cache, context.Background(), model.DNS{IP: "1.1.1.1"}, "example.com", resolve); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected per-resolver cache keys, got %d calls", got)
	}
}

func TestDNSCacheExpiresAtServerTTL(t *testing.T) {
	cache := uptime.NewDNSCache()

	var calls int32
	resolve := func(ctx context.Context, host string) ([]uptime.DNSRecord, error) {
		atomic.AddInt32(&calls, 1)
		return []uptime.DNSRecord{{IP: net.ParseIP("192.0.2.1"), TTL: 20 * time.Millisecond}}, nil
	}

	if _, err := uptime.Lookup(cache, context.Background(), model.DNS{}, "example.com", resolve); err != nil {
		t.Fatal(err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, err := uptime.Lookup(cache, context.Background(), model.DNS{}, "example.com", resolve); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected a re-query after the server TTL elapsed, got %d", got)
	}
}

func TestDNSCacheDoesNotCacheZeroTTL(t *testing.T) {
	cache := uptime.NewDNSCache()

	var calls int32
	resolve := func(ctx context.Context, host string) ([]uptime.DNSRecord, error) {
		atomic.AddInt32(&calls, 1)
		return []uptime.DNSRecord{{IP: net.ParseIP("192.0.2.1"), TTL: 0}}, nil
	}

	for i := 0; i < 2; i++ {
		if _, err := uptime.Lookup(cache, context.Background(), model.DNS{}, "example.com", resolve); err != nil {
			t.Fatal(err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected TTL 0 to bypass the cache, got %d calls", got)
	}
}

func TestDNSCacheDoesNotCacheErrors(t *testing.T) {
	cache := uptime.NewDNSCache()

	var calls int32
	failing := func(ctx context.Context, host string) ([]uptime.DNSRecord, error) {
		atomic.AddInt32(&calls, 1)
		return nil, context.DeadlineExceeded
	}

	for i := 0; i < 2; i++ {
		if _, err := uptime.Lookup(cache, context.Background(), model.DNS{}, "down.example", failing); err == nil {
			t.Fatal("expected lookup error")
		}
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected errors to bypass the cache, got %d calls", got)
	}
}

func TestParseResponseReadsAddressesAndTTL(t *testing.T) {
	name := dnsmessage.MustNewName("example.com.")
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{Response: true, RecursionDesired: true},
		Questions: []dnsmessage.Question{
			{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET},
		},
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 300},
				Body:   &dnsmessage.AResource{A: [4]byte{192, 0, 2, 1}},
			},
		},
	}
	packed, err := msg.Pack()
	if err != nil {
		t.Fatal(err)
	}

	records, cname, err := uptime.ParseResponse(packed)
	if err != nil {
		t.Fatal(err)
	}
	if cname != "" {
		t.Fatalf("unexpected cname %q", cname)
	}
	if len(records) != 1 || !records[0].IP.Equal(net.ParseIP("192.0.2.1")) || records[0].TTL != 300*time.Second {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func TestParseResponseFollowsCNAME(t *testing.T) {
	name := dnsmessage.MustNewName("www.example.com.")
	target := dnsmessage.MustNewName("example.com.")
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{Response: true},
		Questions: []dnsmessage.Question{
			{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET},
		},
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{Name: name, Type: dnsmessage.TypeCNAME, Class: dnsmessage.ClassINET, TTL: 60},
				Body:   &dnsmessage.CNAMEResource{CNAME: target},
			},
		},
	}
	packed, err := msg.Pack()
	if err != nil {
		t.Fatal(err)
	}

	records, cname, err := uptime.ParseResponse(packed)
	if err != nil {
		t.Fatal(err)
	}
	if cname != "example.com" {
		t.Fatalf("expected cname example.com, got %q", cname)
	}
	if len(records) != 0 {
		t.Fatalf("expected no addresses alongside a CNAME, got %+v", records)
	}
}
