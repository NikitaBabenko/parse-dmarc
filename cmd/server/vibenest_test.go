package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmarcguardhq/parse-dmarc/internal/config"
	"github.com/dmarcguardhq/parse-dmarc/internal/logger"
	"github.com/dmarcguardhq/parse-dmarc/internal/pilot"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

// Run inside the Linux image build: Maildir's :2,S suffix is POSIX-specific.
func TestVibeNestSyntheticIngestionPersistence(t *testing.T) {
	log = logger.NewLogger("error", false)
	root := t.TempDir()
	maildir := filepath.Join(root, "maildir")
	db := filepath.Join(root, "reports.db")
	if err := pilot.Seed(maildir); err != nil {
		t.Fatal(err)
	}
	store, err := storage.NewStorage(db)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Maildir: config.MaildirConfig{Path: maildir, DeadletterMailbox: "DEADLETTER"}}
	if err := fetchReports(cfg, store, nil); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = storage.NewStorage(db)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	stats, err := store.GetStatistics()
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalReports != 1 || stats.TotalMessages != 10 || stats.CompliantMessages != 8 || stats.UniqueSourceIPs != 2 {
		t.Fatalf("unexpected persisted statistics: %+v", stats)
	}
	if err := pilot.Seed(maildir); err != nil {
		t.Fatal(err)
	}
	left, err := os.ReadDir(filepath.Join(maildir, "new"))
	if err != nil || len(left) != 0 {
		t.Fatal("fixture reseeded after consumption")
	}
	if err := fetchReports(cfg, store, nil); err != nil {
		t.Fatal(err)
	}
	stats, err = store.GetStatistics()
	if err != nil || stats.TotalReports != 1 || stats.TotalMessages != 10 {
		t.Fatal("repeat ingestion changed counts")
	}
}
