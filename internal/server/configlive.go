package server

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
)

// configLive reloads config from disk when the file mtime advances.
// When the tracked project set, view membership, or local.roots change, it clears
// forge / origin-fetch / scan TTL caches so the board reflects edits without a restart.
type configLive struct {
	mu          sync.Mutex
	path        string
	mod         time.Time
	doc         config.File
	poll        int
	clearCaches func()
	log         observability.Logger
}

func newConfigLive(path string, doc config.File, poll int, clearCaches func(), log observability.Logger) *configLive {
	if log == nil {
		log = observability.NewLogger()
	}
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
func (c *configLive) snapshot() (config.File, int, error) {
	if c == nil {
		return config.File{}, config.DefaultPollSeconds, fmt.Errorf("config unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshotLocked()
}

func (c *configLive) snapshotLocked() (config.File, int, error) {
	if c.path == "" {
		return c.doc, c.poll, nil
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return config.File{}, 0, fmt.Errorf("config stat: %w", err)
	}
	mod := st.ModTime()
	if !mod.After(c.mod) {
		return c.doc, c.poll, nil
	}
	doc, err := config.Load(c.path)
	if err != nil {
		return config.File{}, 0, fmt.Errorf("config reload: %w", err)
	}
	if projectsChanged(c.doc.Projects, doc.Projects) || viewsChanged(c.doc.Views, doc.Views) || localRootsChanged(c.doc.Local.Roots, doc.Local.Roots) {
		if c.clearCaches != nil {
			c.clearCaches()
		}
	}
	c.doc = doc
	c.mod = mod
	c.poll = doc.EffectivePollSeconds()
	return c.doc, c.poll, nil
}

// writable reports whether the live config can be saved to disk.
func (c *configLive) writable() bool {
	return c != nil && strings.TrimSpace(c.path) != ""
}

// replace saves doc to the config path and updates the in-memory snapshot immediately.
func (c *configLive) replace(doc config.File) error {
	if c == nil {
		return fmt.Errorf("config unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.TrimSpace(c.path) == "" {
		return fmt.Errorf("config path not set")
	}
	prev := c.doc
	if err := config.Save(c.path, doc); err != nil {
		return fmt.Errorf("configlive.replace save: %w", err)
	}
	apply := doc
	loaded, err := config.Load(c.path)
	if err != nil {
		c.log.With("path", c.path).Errorf("config reload after save failed: %v (applying saved doc)", err)
	} else {
		apply = loaded
	}
	st, sterr := os.Stat(c.path)
	mod := time.Time{}
	if sterr == nil {
		mod = st.ModTime()
	} else {
		c.log.With("path", c.path).Errorf("config stat after save failed: %v", sterr)
	}
	c.applyLocked(prev, apply, mod)
	return nil
}

func (c *configLive) applyLocked(prev, apply config.File, mod time.Time) {
	if projectsChanged(prev.Projects, apply.Projects) || viewsChanged(prev.Views, apply.Views) || localRootsChanged(prev.Local.Roots, apply.Local.Roots) {
		if c.clearCaches != nil {
			c.clearCaches()
		}
	}
	c.doc = apply
	if !mod.IsZero() {
		c.mod = mod
	}
	c.poll = apply.EffectivePollSeconds()
}

func projectsChanged(a, b []config.Project) bool {
	if len(a) != len(b) {
		return true
	}
	seen := make(map[string]struct{}, len(a))
	for _, p := range a {
		seen[projectFingerprint(p)] = struct{}{}
	}
	for _, p := range b {
		if _, ok := seen[projectFingerprint(p)]; !ok {
			return true
		}
	}
	return false
}

func viewsChanged(a, b []config.View) bool {
	if len(a) != len(b) {
		return true
	}
	seen := make(map[string]struct{}, len(a))
	for _, v := range a {
		seen[viewFingerprint(v)] = struct{}{}
	}
	for _, v := range b {
		if _, ok := seen[viewFingerprint(v)]; !ok {
			return true
		}
	}
	return false
}

func localRootsChanged(a, b []string) bool {
	if len(a) != len(b) {
		return true
	}
	seen := make(map[string]struct{}, len(a))
	for _, r := range a {
		seen[r] = struct{}{}
	}
	for _, r := range b {
		if _, ok := seen[r]; !ok {
			return true
		}
	}
	return false
}

func projectFingerprint(p config.Project) string {
	return string(p.Host) + "\x00" + p.Path + "\x00" + p.ID + "\x00" + p.LocalPath
}

func viewFingerprint(v config.View) string {
	return v.ID + "\x00" + v.Label + "\x00" + strings.Join(v.Projects, "\x01")
}
