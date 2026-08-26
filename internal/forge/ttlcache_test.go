package forge

import (
	"errors"
	"testing"
	"time"
)

func TestTTLCacheHeadsHitMissFresh(t *testing.T) {
	cache := NewTTLCache()
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	cache.SetNow(func() time.Time { return now })

	loads := 0
	load := func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{
			DefaultBranch: "main",
			Heads:         []RemoteHead{{Name: "main"}, {Name: "feat/a"}},
		}, nil
	}

	ttl := 2 * time.Minute
	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, false, load); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, false, load); err != nil {
		t.Fatal(err)
	}
	if loads != 1 {
		t.Fatalf("want 1 load after hit, got %d", loads)
	}

	now = now.Add(3 * time.Minute)
	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, false, load); err != nil {
		t.Fatal(err)
	}
	if loads != 2 {
		t.Fatalf("want 2 loads after expiry, got %d", loads)
	}

	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, true, load); err != nil {
		t.Fatal(err)
	}
	if loads != 3 {
		t.Fatalf("want 3 loads after fresh, got %d", loads)
	}
}

func TestTTLCacheMergedFailureClearsOK(t *testing.T) {
	cache := NewTTLCache()
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	cache.SetNow(func() time.Time { return now })

	okLoad := func() ([]MergedReview, error) {
		return []MergedReview{{Branch: "feat/x", ID: 1}}, nil
	}
	list, ok := cache.GetOrLoadMerged("github/o/r", 10*time.Minute, false, okLoad)
	if !ok || len(list) != 1 {
		t.Fatalf("first load: ok=%v list=%v", ok, list)
	}

	failLoad := func() ([]MergedReview, error) {
		return nil, errors.New("boom")
	}
	now = now.Add(11 * time.Minute)
	list, ok = cache.GetOrLoadMerged("github/o/r", 10*time.Minute, false, failLoad)
	if ok || list != nil {
		t.Fatalf("expired fail must clear MergedOK: ok=%v list=%v", ok, list)
	}
}

func TestTTLCacheMergedZeroTTLAlwaysLoads(t *testing.T) {
	cache := NewTTLCache()
	loads := 0
	load := func() ([]MergedReview, error) {
		loads++
		return []MergedReview{{Branch: "a"}}, nil
	}
	cache.GetOrLoadMerged("k", 0, false, load)
	cache.GetOrLoadMerged("k", 0, false, load)
	if loads != 2 {
		t.Fatalf("zero ttl should always load, got %d", loads)
	}
}

func TestTTLCacheClear(t *testing.T) {
	cache := NewTTLCache()
	loads := 0
	load := func() (HeadsSnapshot, error) {
		loads++
		return HeadsSnapshot{DefaultBranch: "main"}, nil
	}
	ttl := time.Minute
	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, false, load); err != nil {
		t.Fatal(err)
	}
	cache.Clear()
	if _, err := cache.GetOrLoadHeads("github/o/r", ttl, false, load); err != nil {
		t.Fatal(err)
	}
	if loads != 2 {
		t.Fatalf("clear should force reload, got %d loads", loads)
	}
	cache.Clear() // nil-safe and empty-safe
	(*TTLCache)(nil).Clear()
}
