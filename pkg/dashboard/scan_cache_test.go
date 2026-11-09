package dashboard

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

type countingScanLocal struct {
	scans atomic.Int32
	syncLocalFake
}

func (c *countingScanLocal) ScanRoots(context.Context, []string) localgit.Discovery {
	c.scans.Add(1)
	return localgit.Discovery{ByKey: map[string][]localgit.Checkout{}}
}

func TestScanRootsCachedRespectsTTLAndFresh(t *testing.T) {
	fx := dualHostFake()
	local := &countingScanLocal{}
	now := time.Date(2026, 11, 8, 9, 0, 0, 0, time.UTC)
	cfg := testConfig(fx, local)
	cfg.Clock = func() time.Time { return now }
	s := cfg.CreateService()
	doc := boardconfig.File{
		Projects: []boardconfig.Project{
			{ID: "gh-app", Label: "App", Host: boardconfig.HostGitHub, Path: "acme/app"},
		},
		Local: boardconfig.Local{Roots: []string{"/tmp/gitboard-scan-cache"}},
	}

	if _, err := s.Collect(context.Background(), doc, false, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Collect(context.Background(), doc, false, ""); err != nil {
		t.Fatal(err)
	}
	if got := local.scans.Load(); got != 1 {
		t.Fatalf("ttl hit: scans=%d want 1", got)
	}

	if _, err := s.Collect(context.Background(), doc, true, ""); err != nil {
		t.Fatal(err)
	}
	if got := local.scans.Load(); got != 2 {
		t.Fatalf("fresh bypass: scans=%d want 2", got)
	}

	s.ClearCaches()
	if _, err := s.Collect(context.Background(), doc, false, ""); err != nil {
		t.Fatal(err)
	}
	if got := local.scans.Load(); got != 3 {
		t.Fatalf("after clear: scans=%d want 3", got)
	}

	// Expired TTL forces a rescan (default scan TTL is 300s).
	now = now.Add(6 * time.Minute)
	if _, err := s.Collect(context.Background(), doc, false, ""); err != nil {
		t.Fatal(err)
	}
	if got := local.scans.Load(); got != 4 {
		t.Fatalf("expired ttl: scans=%d want 4", got)
	}
}
