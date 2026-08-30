package remotegit

import (
	"testing"
	"time"
)

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
		return HeadsSnapshot{}, nil
	})
	c.GetOrLoadHeads("k", time.Hour, true, func() (HeadsSnapshot, error) { //nolint
		loads++
		return HeadsSnapshot{}, nil
	})
	if loads != 2 {
		t.Fatalf("fresh must bypass: loads=%d", loads)
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
