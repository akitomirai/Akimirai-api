package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestPgDumperDumpMissingPgDumpReturnsActionableError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dumper := &PgDumper{cfg: &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "sub2api",
		DBName: "sub2api",
	}}

	reader, err := dumper.Dump(context.Background())
	if reader != nil {
		_ = reader.Close()
	}

	if err == nil {
		t.Fatal("expected missing pg_dump error")
	}
	msg := err.Error()
	for _, want := range []string{
		"pg_dump not found in PATH",
		"install PostgreSQL client tools",
		"bundled pg_dump/psql",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected error %q to contain %q", msg, want)
		}
	}
}

func TestPgDumperRestoreMissingPsqlReturnsActionableError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dumper := &PgDumper{cfg: &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "sub2api",
		DBName: "sub2api",
	}}

	err := dumper.Restore(context.Background(), strings.NewReader("-- dump"))
	if err == nil {
		t.Fatal("expected missing psql error")
	}
	msg := err.Error()
	for _, want := range []string{
		"psql not found in PATH",
		"install PostgreSQL client tools",
		"bundled pg_dump/psql",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected error %q to contain %q", msg, want)
		}
	}
}
