package server

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// configLive reloads config from disk when the file mtime advances.
// When the tracked project set, view membership, or local.roots change, it clears
// forge / origin-fetch / scan TTL caches so the board reflects edits without a restart.
type configLive struct {
	mu          sync.Mutex
	path        string
	mod         time.Time
	doc         boardconfig.File
	poll        int
	clearCaches func()
	log         observability.Logger
}

// newConfigLive builds the live config. log is required; NewMux checks it.
func newConfigLive(path string, doc boardconfig.File, poll int, clearCaches func(), log observability.Logger) *configLive {
	cl := &configLive{
		path:        path,
		doc:         doc,
		poll:        poll,
		clearCaches: clearCaches,
		log:         log,
	}
	if path != "" {
		if st, err := os.Stat(path); err == nil {
			cl.mod = st.ModTime()
		}
	}
	return cl
}

// snapshot returns the current config. When ConfigPath is set and the file
// mtime is newer, it reloads the full File from disk. Reload failure fails
// closed (error) so callers do not keep serving a silent stale snapshot.
func (c *configLive) snapshot() (boardconfig.File, int, error) {
	if c == nil {
		return boardconfig.File{}, config.DefaultPollSeconds, fmt.Errorf("config unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshotLocked()
}

func (c *configLive) snapshotLocked() (boardconfig.File, int, error) {
	if c.path == "" {
		return c.doc, c.poll, nil
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return boardconfig.File{}, 0, fmt.Errorf("config stat: %w", err)
	}
	mod := st.ModTime()
	if !mod.After(c.mod) {
		return c.doc, c.poll, nil
	}
	doc, err := config.Load(c.path)
	if err != nil {
		return boardconfig.File{}, 0, fmt.Errorf("config reload: %w", err)
	}
	c.applyLocked(c.doc, doc, mod)
	return c.doc, c.poll, nil
}

// writable reports whether the live config can be saved to disk.
func (c *configLive) writable() bool {
	return c != nil && strings.TrimSpace(c.path) != ""
}

// replace saves doc to the config path and updates the in-memory snapshot immediately.
func (c *configLive) replace(doc boardconfig.File) error {
	if c == nil {
		return fmt.Errorf("config unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.TrimSpace(c.path) == "" {
		return fmt.Errorf("config path not set")
	}
	const op = "server.configLive.replace"
	prev := c.doc
	if err := config.Save(c.path, doc); err != nil {
		return gberr.Wrap(err, gberr.CodeUnavailable, op, "saving the config file failed").With("path", c.path)
	}
	// Read the file back: memory must hold what the next process start will
	// load. If the saved file cannot be reloaded, keep the previous snapshot
	// and fail the request; the next snapshot retries the reload by mtime.
	loaded, err := config.Load(c.path)
	if err != nil {
		return gberr.Wrap(err, gberr.CodeInternal, op, "config saved but reload failed").With("path", c.path)
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return gberr.Wrap(err, gberr.CodeInternal, op, "config saved but stat failed").With("path", c.path)
	}
	c.applyLocked(prev, loaded, st.ModTime())
	return nil
}

// applyLocked installs apply as the live snapshot and clears the TTL caches
// when a cache-relevant field changed.
func (c *configLive) applyLocked(prev, apply boardconfig.File, mod time.Time) {
	if cacheRelevantChange(prev, apply) && c.clearCaches != nil {
		c.clearCaches()
	}
	c.doc = apply
	if !mod.IsZero() {
		c.mod = mod
	}
	c.poll = config.EffectivePollSeconds(apply)
}

// cacheRelevantChange reports whether the tracked projects, the views or the
// local roots differ. Those feed the forge, origin-fetch and scan caches.
func cacheRelevantChange(prev, next boardconfig.File) bool {
	return setChanged(prev.Projects, next.Projects, projectFingerprint) ||
		setChanged(prev.Views, next.Views, viewFingerprint) ||
		setChanged(prev.Local.Roots, next.Local.Roots, func(s string) string { return s })
}

// setChanged reports whether a and b differ as sets of key(item). Equal
// lengths with the same members count as unchanged, in any order.
func setChanged[T any](a, b []T, key func(T) string) bool {
	if len(a) != len(b) {
		return true
	}
	seen := make(map[string]struct{}, len(a))
	for _, v := range a {
		seen[key(v)] = struct{}{}
	}
	for _, v := range b {
		if _, ok := seen[key(v)]; !ok {
			return true
		}
	}
	return false
}

func projectFingerprint(p boardconfig.Project) string {
	return string(p.Host) + "\x00" + p.Path + "\x00" + p.ID + "\x00" + p.LocalPath
}

func viewFingerprint(v boardconfig.View) string {
	return v.ID + "\x00" + v.Label + "\x00" + strings.Join(v.Projects, "\x01")
}
