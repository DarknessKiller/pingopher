package uptime

import (
	"context"
	"net"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

// TestQueryServerPacketIsWellFormed guards against the builder prefixing the
// query with the backing array's zero bytes: the fake server rejects anything
// that is not exactly one A question, so a padded packet fails as FORMERR.
func TestQueryServerPacketIsWellFormed(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 4096)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}

		name := dnsmessage.MustNewName("example.com.")
		var parser dnsmessage.Parser
		header, perr := parser.Start(buf[:n])
		questions, qerr := parser.AllQuestions()
		valid := perr == nil && qerr == nil && len(questions) == 1 &&
			questions[0].Name == name && questions[0].Type == dnsmessage.TypeA

		resp := dnsmessage.Message{Header: dnsmessage.Header{
			ID:                 header.ID,
			Response:           true,
			RecursionDesired:   true,
			RecursionAvailable: true,
		}}
		if !valid {
			resp.Header.RCode = dnsmessage.RCodeFormatError
		} else {
			resp.Questions = []dnsmessage.Question{questions[0]}
			resp.Answers = []dnsmessage.Resource{{
				Header: dnsmessage.ResourceHeader{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 60},
				Body:   &dnsmessage.AResource{A: [4]byte{192, 0, 2, 1}},
			}}
		}
		packed, err := resp.Pack()
		if err != nil {
			return
		}
		_, _ = conn.WriteTo(packed, addr)
	}()

	records, _, err := queryServer(context.Background(), conn.LocalAddr().String(), "udp", "example.com.", dnsmessage.TypeA)
	if err != nil {
		t.Fatalf("queryServer: %v", err)
	}
	if len(records) != 1 || !records[0].ip.Equal(net.ParseIP("192.0.2.1")) {
		t.Fatalf("unexpected records: %+v", records)
	}
}
