package dashboard

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
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
	s := New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), local)
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
		},
		Local: config.Local{Roots: []string{"/tmp/gitboard-scan-cache"}},
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
	s.scanMu.Lock()
	s.scanAt = time.Now().Add(-6 * time.Minute)
	s.scanMu.Unlock()
	if _, err := s.Collect(context.Background(), doc, false, ""); err != nil {
		t.Fatal(err)
	}
	if got := local.scans.Load(); got != 4 {
		t.Fatalf("expired ttl: scans=%d want 4", got)
	}
}
