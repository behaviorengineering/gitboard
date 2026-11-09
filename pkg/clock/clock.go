// Package clock is public API: the injectable time source behind every
// gitboard timestamp and TTL (dashboard stamps, scan and forge caches,
// origin-fetch freshness, prune session records).
package clock

import "time"

// Clock returns the current time; tests pass a fixed or stepping clock.
type Clock func() time.Time

// OrWall returns c, or time.Now when c is nil. Only service constructors at
// the top of a dependency tree call it; caches and leaf helpers require a
// non-nil Clock.
func OrWall(c Clock) Clock {
	if c == nil {
		return time.Now
	}
	return c
}
