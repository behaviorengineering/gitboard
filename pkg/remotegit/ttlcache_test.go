package remotegit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

var errMergedLoad = errors.New("merged load failed")

func TestTTLCacheHeadsHit(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	now := fixed
	c := NewTTLCache(func() time.Time { return now })

	loads := 0
	snap, err := c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil || snap.DefaultBranch != "main" || loads != 1 {
		t.Fatalf("first load: %v loads=%d", err, loads)
	}

	snap, err = c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil || snap.DefaultBranch != "main" || loads != 1 {
		t.Fatalf("cache hit: %v loads=%d", err, loads)
	}

	now = fixed.Add(2 * time.Minute)
	snap, err = c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{DefaultBranch: "develop"}, nil
	})
	if err != nil || snap.DefaultBranch != "develop" || loads != 2 {
		t.Fatalf("expired: %v loads=%d branch=%s", err, loads, snap.DefaultBranch)
	}
}

func TestTTLCacheFreshBypassesCache(t *testing.T) {
	c := NewTTLCache(time.Now)
	loads := 0
	c.GetOrLoadHeads("k", time.Hour, false, func() (HeadsSnapshot, error) { //nolint
		loads++
		return HeadsSnapshot{DefaultBranch: "old"}, nil
	})
	c.GetOrLoadHeads("k", time.Hour, true, func() (HeadsSnapshot, error) { //nolint
		loads++
		return HeadsSnapshot{DefaultBranch: "new"}, nil
	})
	if loads != 2 {
		t.Fatalf("fresh must bypass: loads=%d", loads)
	}
	snap, err := c.GetOrLoadHeads("k", time.Hour, false, func() (HeadsSnapshot, error) {
		t.Fatal("fresh success must write cache")
		return HeadsSnapshot{}, nil
	})
	if err != nil || snap.DefaultBranch != "new" {
		t.Fatalf("fresh write-back: %v %+v", err, snap)
	}
}

func TestTTLCacheMergedFreshWriteBack(t *testing.T) {
	c := NewTTLCache(time.Now)
	loads := 0
	if _, err := c.GetOrLoadMerged("k", time.Hour, false, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "old", ID: 1}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	list, err := c.GetOrLoadMerged("k", time.Hour, true, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/new", ID: 2}}, nil
	})
	if err != nil || loads != 2 || len(list) != 1 || list[0].Branch != "feat/new" {
		t.Fatalf("fresh merged: err=%v loads=%d %+v", err, loads, list)
	}
	list, err = c.GetOrLoadMerged("k", time.Hour, false, func() ([]board.MergedReview, error) {
		t.Fatal("fresh merged success must write cache")
		return nil, nil
	})
	if err != nil || len(list) != 1 || list[0].ID != 2 {
		t.Fatalf("merged write-back: err=%v %+v", err, list)
	}
}

func TestTTLCacheMergedLoadErrorClearsOK(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	now := fixed
	c := NewTTLCache(func() time.Time { return now })
	if _, err := c.GetOrLoadMerged("k", time.Minute, false, func() ([]board.MergedReview, error) {
		return []board.MergedReview{{Branch: "feat/a", ID: 1}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	now = fixed.Add(2 * time.Minute)
	list, err := c.GetOrLoadMerged("k", time.Minute, false, func() ([]board.MergedReview, error) {
		return nil, errMergedLoad
	})
	if err == nil || list != nil {
		t.Fatalf("expired failed refresh must not keep MergedOK: err=%v %+v", err, list)
	}
}

func TestTTLCacheMergedBranchPositiveSticky(t *testing.T) {
	c := NewTTLCache(time.Now)
	loads := 0
	list, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/a", ID: 9}}, nil
	})
	if err != nil || loads != 1 || list[0].ID != 9 {
		t.Fatalf("first branch load: err=%v loads=%d %+v", err, loads, list)
	}
	list, err = c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		t.Fatal("positive must not expire")
		return nil, nil
	})
	if err != nil || list[0].ID != 9 {
		t.Fatalf("sticky positive: err=%v %+v", err, list)
	}
	c.ForgetMergedBranch("k", "feat/a")
	list, err = c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		loads++
		return nil, nil
	})
	if err != nil || list != nil || loads != 2 {
		t.Fatalf("after forget: err=%v loads=%d %+v", err, loads, list)
	}
}

func TestTTLCacheMergedBranchNegativeTTL(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	now := fixed
	c := NewTTLCache(func() time.Time { return now })
	loads := 0
	_, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/gone", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		loads++
		return nil, nil
	})
	if err != nil || loads != 1 {
		t.Fatalf("negative store: err=%v loads=%d", err, loads)
	}
	_, err = c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/gone", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		t.Fatal("negative still in TTL")
		return nil, nil
	})
	if err != nil {
		t.Fatal("negative hit")
	}
	now = fixed.Add(2 * time.Minute)
	list, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/gone", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/gone", ID: 3}}, nil
	})
	if err != nil || loads != 2 || list[0].ID != 3 {
		t.Fatalf("negative expired: err=%v loads=%d %+v", err, loads, list)
	}
}

func TestTTLCacheMergedBranchErrorKeepsPositive(t *testing.T) {
	c := NewTTLCache(time.Now)
	if _, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: false}, func() ([]board.MergedReview, error) {
		return []board.MergedReview{{Branch: "feat/a", ID: 4}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	list, err := c.GetOrLoadMergedBranch(MergedBranchQuery{Key: "k", Branch: "feat/a", NegativeTTL: time.Minute, Fresh: true}, func() ([]board.MergedReview, error) {
		return nil, errMergedLoad
	})
	if err != nil || len(list) != 1 || list[0].ID != 4 {
		t.Fatalf("error must keep positive: err=%v %+v", err, list)
	}
}

func TestTTLCacheClear(t *testing.T) {
	c := NewTTLCache(time.Now)
	loads := 0
	c.GetOrLoadHeads("k", time.Hour, false, func() (HeadsSnapshot, error) { //nolint
		loads++
		return HeadsSnapshot{}, nil
	})
	c.Clear()
	c.GetOrLoadHeads("k", time.Hour, false, func() (HeadsSnapshot, error) { //nolint
		loads++
		return HeadsSnapshot{}, nil
	})
	if loads != 2 {
		t.Fatalf("after clear must reload: loads=%d", loads)
	}
}

func TestTTLCacheNilSafe(t *testing.T) {
	var c *TTLCache
	c.Clear()
	_, err := c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		return HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil {
		t.Fatalf("nil cache should load: %v", err)
	}
}

func TestNewTTLCacheRequiresClock(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic for nil clock")
		}
	}()
	NewTTLCache(nil)
}

func TestPageLoopStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pages := 0
	err := pageLoop(ctx, func(int) (bool, error) {
		pages++
		cancel()
		return false, nil
	})
	if pages != 1 || gberr.CodeOf(err) != gberr.CodeCanceled {
		t.Fatalf("pages=%d err=%v", pages, err)
	}
}

func TestNewRepoRefChecksShapeAndNamesFromPath(t *testing.T) {
	if _, ok := newRepoRef(boardconfig.HostAzureDevOps, "org/repo", ""); ok {
		t.Fatal("azure devops ref needs three segments")
	}
	ref, ok := newRepoRef(boardconfig.HostGitLab, "/g/sub/app/", "")
	if !ok || ref.Path != "g/sub/app" || ref.Name != "app" {
		t.Fatalf("ref=%+v ok=%v", ref, ok)
	}
}

// TestCacheKeyTrimsPath keeps the loader's key and InvalidateProject's key
// equal when the configured path carries spaces or slashes.
func TestCacheKeyTrimsPath(t *testing.T) {
	want := CacheKey(boardconfig.HostGitHub, "acme/app")
	for _, path := range []string{" acme/app ", "/acme/app/", " /acme/app/ "} {
		if got := CacheKey(boardconfig.HostGitHub, path); got != want {
			t.Errorf("CacheKey(%q)=%q want %q", path, got, want)
		}
	}
}
