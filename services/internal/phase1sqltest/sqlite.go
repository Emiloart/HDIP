package phase1sqltest

import (
	"context"
	"path/filepath"
	"testing"

	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
	_ "modernc.org/sqlite"
)

func SQLiteDSN(t testing.TB) string {
	t.Helper()

	return "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "phase1.sqlite")) + "?mode=rwc&cache=shared"
}

func OpenSQLiteStore(t testing.TB) *phase1sql.Store {
	t.Helper()

	return OpenSQLiteStoreAtDSN(t, SQLiteDSN(t))
}

func OpenSQLiteStoreAtDSN(t testing.TB, dsn string) *phase1sql.Store {
	t.Helper()

	if err := phase1sql.MigrateUp(context.Background(), "sqlite", dsn); err != nil {
		t.Fatalf("migrate sqlite store: %v", err)
	}

	store, err := phase1sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}
