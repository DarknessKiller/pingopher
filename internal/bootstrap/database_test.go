package bootstrap_test

import (
	"testing"

	"github.com/DarknessKiller/pingopher/internal/bootstrap"
)

func TestSQLiteDSNPragmas(t *testing.T) {
	got := bootstrap.SQLiteDSN("pingopher.db")
	want := "pingopher.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if bootstrap.SQLiteDSN("file:x?y") != "file:x?y&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)" {
		t.Error("existing query separator must survive")
	}
}
