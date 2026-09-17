package remotegit

import (
	"errors"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/board"
)

var errMergedLoad = errors.New("merged load failed")

func TestTTLCacheHeadsHit(t *testing.T) {
	c := NewTTLCache()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c.SetNow(func() time.Time { return fixed })

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

	c.SetNow(func() time.Time { return fixed.Add(2 * time.Minute) })
	snap, err = c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{DefaultBranch: "develop"}, nil
	})
	if err != nil || snap.DefaultBranch != "develop" || loads != 2 {
		t.Fatalf("expired: %v loads=%d branch=%s", err, loads, snap.DefaultBranch)
	}
}

func TestTTLCacheFreshBypassesCache(t *testing.T) {
	c := NewTTLCache()
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
	c := NewTTLCache()
	loads := 0
	c.GetOrLoadMerged("k", time.Hour, false, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "old", ID: 1}}, nil
	})
	list, ok := c.GetOrLoadMerged("k", time.Hour, true, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/new", ID: 2}}, nil
	})
	if !ok || loads != 2 || len(list) != 1 || list[0].Branch != "feat/new" {
		t.Fatalf("fresh merged: ok=%v loads=%d %+v", ok, loads, list)
	}
	list, ok = c.GetOrLoadMerged("k", time.Hour, false, func() ([]board.MergedReview, error) {
		t.Fatal("fresh merged success must write cache")
		return nil, nil
	})
	if !ok || len(list) != 1 || list[0].ID != 2 {
		t.Fatalf("merged write-back: ok=%v %+v", ok, list)
	}
}

func TestTTLCacheMergedLoadErrorClearsOK(t *testing.T) {
	c := NewTTLCache()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c.SetNow(func() time.Time { return fixed })
	c.GetOrLoadMerged("k", time.Minute, false, func() ([]board.MergedReview, error) {
		return []board.MergedReview{{Branch: "feat/a", ID: 1}}, nil
	})
	c.SetNow(func() time.Time { return fixed.Add(2 * time.Minute) })
	list, ok := c.GetOrLoadMerged("k", time.Minute, false, func() ([]board.MergedReview, error) {
		return nil, errMergedLoad
	})
	if ok || list != nil {
		t.Fatalf("expired failed refresh must not keep MergedOK: ok=%v %+v", ok, list)
	}
}

func TestTTLCacheMergedBranchPositiveSticky(t *testing.T) {
	c := NewTTLCache()
	loads := 0
	list, ok := c.GetOrLoadMergedBranch("k", "feat/a", time.Minute, false, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/a", ID: 9}}, nil
	})
	if !ok || loads != 1 || list[0].ID != 9 {
		t.Fatalf("first branch load: ok=%v loads=%d %+v", ok, loads, list)
	}
	list, ok = c.GetOrLoadMergedBranch("k", "feat/a", time.Minute, false, func() ([]board.MergedReview, error) {
		t.Fatal("positive must not expire")
		return nil, nil
	})
	if !ok || list[0].ID != 9 {
		t.Fatalf("sticky positive: ok=%v %+v", ok, list)
	}
	c.ForgetMergedBranch("k", "feat/a")
	list, ok = c.GetOrLoadMergedBranch("k", "feat/a", time.Minute, false, func() ([]board.MergedReview, error) {
		loads++
		return nil, nil
	})
	if !ok || list != nil || loads != 2 {
		t.Fatalf("after forget: ok=%v loads=%d %+v", ok, loads, list)
	}
}

func TestTTLCacheMergedBranchNegativeTTL(t *testing.T) {
	c := NewTTLCache()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c.SetNow(func() time.Time { return fixed })
	loads := 0
	_, ok := c.GetOrLoadMergedBranch("k", "feat/gone", time.Minute, false, func() ([]board.MergedReview, error) {
		loads++
		return nil, nil
	})
	if !ok || loads != 1 {
		t.Fatalf("negative store: ok=%v loads=%d", ok, loads)
	}
	_, ok = c.GetOrLoadMergedBranch("k", "feat/gone", time.Minute, false, func() ([]board.MergedReview, error) {
		t.Fatal("negative still in TTL")
		return nil, nil
	})
	if !ok {
		t.Fatal("negative hit")
	}
	c.SetNow(func() time.Time { return fixed.Add(2 * time.Minute) })
	list, ok := c.GetOrLoadMergedBranch("k", "feat/gone", time.Minute, false, func() ([]board.MergedReview, error) {
		loads++
		return []board.MergedReview{{Branch: "feat/gone", ID: 3}}, nil
	})
	if !ok || loads != 2 || list[0].ID != 3 {
		t.Fatalf("negative expired: ok=%v loads=%d %+v", ok, loads, list)
	}
}

func TestTTLCacheMergedBranchErrorKeepsPositive(t *testing.T) {
	c := NewTTLCache()
	c.GetOrLoadMergedBranch("k", "feat/a", time.Minute, false, func() ([]board.MergedReview, error) {
		return []board.MergedReview{{Branch: "feat/a", ID: 4}}, nil
	})
	list, ok := c.GetOrLoadMergedBranch("k", "feat/a", time.Minute, true, func() ([]board.MergedReview, error) {
		return nil, errMergedLoad
	})
	if !ok || len(list) != 1 || list[0].ID != 4 {
		t.Fatalf("error must keep positive: ok=%v %+v", ok, list)
	}
}

func TestTTLCacheClear(t *testing.T) {
	c := NewTTLCache()
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
	c.SetNow(nil)
	c.Clear()
	_, err := c.GetOrLoadHeads("k", time.Minute, false, func() (HeadsSnapshot, error) {
		return HeadsSnapshot{DefaultBranch: "main"}, nil
	})
	if err != nil {
		t.Fatalf("nil cache should load: %v", err)
	}
}
