package uptime

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/DarknessKiller/pingopher/internal/model"
)

const dnsQueryTimeout = 2 * time.Second

type dnsRecord struct {
	IP  net.IP
	TTL time.Duration
}

// resolveFunc performs one uncached lookup for host.
type resolveFunc func(ctx context.Context, host string) ([]dnsRecord, error)

// resolveFuncFor reads the nameserver list lazily, so it is only touched on a cache miss.
func resolveFuncFor(dns model.DNS) resolveFunc {
	return func(ctx context.Context, host string) ([]dnsRecord, error) {
		servers, network := nameServersFor(dns)
		return resolveHost(ctx, servers, network, host)
	}
}

// System DNS reads /etc/resolv.conf; a custom entry targets its own configured address.
func nameServersFor(dns model.DNS) ([]string, string) {
	network := strings.ToLower(dns.Protocol)
	if network != "tcp" {
		network = "udp"
	}

	if dns.IP == "" || dns.Name == "System DNS" {
		return systemNameServers(), network
	}

	port := "53"
	if dns.Port != 0 {
		port = strconv.Itoa(int(dns.Port))
	}
	return []string{net.JoinHostPort(dns.IP, port)}, network
}

func systemNameServers() []string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}

	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "nameserver" {
			continue
		}
		servers = append(servers, net.JoinHostPort(fields[1], "53"))
	}
	return servers
}

// resolveHost follows CNAME chains and carries each record's server-assigned TTL back to the caller.
func resolveHost(ctx context.Context, servers []string, network, host string) ([]dnsRecord, error) {
	if len(servers) == 0 {
		return nil, errors.New("no DNS nameservers available")
	}

	lookup := fqdn(host)
	seen := make(map[string]bool, 4)

	for i := 0; i < 8 && !seen[lookup]; i++ {
		seen[lookup] = true

		var records []dnsRecord
		cname := ""
		for _, qtype := range []dnsmessage.Type{dnsmessage.TypeA, dnsmessage.TypeAAAA} {
			recs, cn, err := queryServers(ctx, servers, network, lookup, qtype)
			if err != nil {
				return nil, err
			}
			records = append(records, recs...)
			if cn != "" {
				cname = cn
			}
		}

		if len(records) > 0 {
			return records, nil
		}
		if cname == "" {
			break
		}
		lookup = fqdn(cname)
	}

	return nil, fmt.Errorf("no addresses found for %s", host)
}

func queryServers(ctx context.Context, servers []string, network, host string, qtype dnsmessage.Type) ([]dnsRecord, string, error) {
	var lastErr error
	for _, server := range servers {
		records, cname, err := queryServer(ctx, server, network, host, qtype)
		if err == nil {
			return records, cname, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no DNS nameservers available")
	}
	return nil, "", lastErr
}

func queryServer(ctx context.Context, server, network, host string, qtype dnsmessage.Type) ([]dnsRecord, string, error) {
	name, err := dnsmessage.NewName(host)
	if err != nil {
		return nil, "", err
	}

	// NewBuilder appends to the slice it is given, so hand it a zero-length
	// slice; passing buf[:] would prefix the packet with len(buf) zero bytes and
	// every server would reject it as malformed (FORMERR).
	var buf [512]byte
	builder := dnsmessage.NewBuilder(buf[:0], dnsmessage.Header{RecursionDesired: true})
	builder.EnableCompression()
	if err := builder.StartQuestions(); err != nil {
		return nil, "", err
	}
	if err := builder.Question(dnsmessage.Question{Name: name, Type: qtype, Class: dnsmessage.ClassINET}); err != nil {
		return nil, "", err
	}
	query, err := builder.Finish()
	if err != nil {
		return nil, "", err
	}

	resp, err := exchange(ctx, server, network, query)
	if err != nil {
		return nil, "", err
	}
	if truncated(resp) {
		if resp, err = exchange(ctx, server, "tcp", query); err != nil {
			return nil, "", err
		}
	}
	return parseResponse(resp)
}

// TCP answers are length-prefixed per RFC 1035; UDP answers are not.
func exchange(ctx context.Context, server, network string, query []byte) ([]byte, error) {
	dialer := net.Dialer{Timeout: dnsQueryTimeout}
	conn, err := dialer.DialContext(ctx, network, server)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	deadline := time.Now().Add(dnsQueryTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	if network == "tcp" {
		framed := make([]byte, 2+len(query))
		binary.BigEndian.PutUint16(framed, uint16(len(query)))
		copy(framed[2:], query)
		if _, err := conn.Write(framed); err != nil {
			return nil, err
		}

		head := make([]byte, 2)
		if _, err := io.ReadFull(conn, head); err != nil {
			return nil, err
		}
		resp := make([]byte, binary.BigEndian.Uint16(head))
		if _, err := io.ReadFull(conn, resp); err != nil {
			return nil, err
		}
		return resp, nil
	}

	if _, err := conn.Write(query); err != nil {
		return nil, err
	}
	resp := make([]byte, 4096)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	return resp[:n], nil
}

func truncated(resp []byte) bool {
	var parser dnsmessage.Parser
	header, err := parser.Start(resp)
	return err == nil && header.Truncated
}

// NXDOMAIN is reported as an empty, error-free answer so a missing name falls through to the
// "no addresses" path.
func parseResponse(resp []byte) ([]dnsRecord, string, error) {
	var parser dnsmessage.Parser
	header, err := parser.Start(resp)
	if err != nil {
		return nil, "", err
	}
	if header.RCode == dnsmessage.RCodeNameError {
		return nil, "", nil
	}
	if header.RCode != dnsmessage.RCodeSuccess {
		return nil, "", fmt.Errorf("dns: %s", header.RCode)
	}
	if err := parser.SkipAllQuestions(); err != nil {
		return nil, "", err
	}

	var records []dnsRecord
	cname := ""
	for {
		ah, err := parser.AnswerHeader()
		if errors.Is(err, dnsmessage.ErrSectionDone) {
			break
		}
		if err != nil {
			return nil, "", err
		}

		ttl := time.Duration(ah.TTL) * time.Second
		switch ah.Type {
		case dnsmessage.TypeA:
			r, err := parser.AResource()
			if err != nil {
				return nil, "", err
			}
			records = append(records, dnsRecord{IP: net.IP(r.A[:]), TTL: ttl})
		case dnsmessage.TypeAAAA:
			r, err := parser.AAAAResource()
			if err != nil {
				return nil, "", err
			}
			records = append(records, dnsRecord{IP: net.IP(r.AAAA[:]), TTL: ttl})
		case dnsmessage.TypeCNAME:
			r, err := parser.CNAMEResource()
			if err != nil {
				return nil, "", err
			}
			cname = strings.TrimSuffix(r.CNAME.String(), ".")
		default:
			if err := parser.SkipAnswer(); err != nil {
				return nil, "", err
			}
		}
	}
	return records, cname, nil
}

func fqdn(host string) string {
	if strings.HasSuffix(host, ".") {
		return host
	}
	return host + "."
}
