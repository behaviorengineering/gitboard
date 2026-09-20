package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
)

// configLive reloads config from disk when the file mtime advances.
// When the tracked project set or view membership changes, it clears forge and
// origin-fetch TTL caches so the board reflects sync/UI edits without a restart.
type configLive struct {
	mu          sync.Mutex
	path        string
	mod         time.Time
	doc         config.File
	poll        int
	clearCaches func()
}

func newConfigLive(path string, doc config.File, poll int, clearCaches func()) *configLive {
	cl := &configLive{
		path:        path,
		doc:         doc,
		poll:        poll,
		clearCaches: clearCaches,
	}
	if path != "" {
		if st, err := os.Stat(path); err == nil {
			cl.mod = st.ModTime()
		}
	}
	return cl
}

// snapshot returns the current config. When ConfigPath is set and the file
// mtime is newer, it reloads the full File from disk. On reload failure it
// keeps the last good snapshot. Project-set or view membership changes clear
// the forge cache.
func (c *configLive) snapshot() (config.File, int) {
	if c == nil {
		return config.File{}, config.DefaultPollSeconds
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshotLocked()
}

func (c *configLive) snapshotLocked() (config.File, int) {
	if c.path == "" {
		return c.doc, c.poll
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return c.doc, c.poll
	}
	mod := st.ModTime()
	if !mod.After(c.mod) {
		return c.doc, c.poll
	}
	doc, err := config.Load(c.path)
	if err != nil {
		log.Printf("gitboard: config reload failed path=%s: %v", c.path, err)
		return c.doc, c.poll
	}
	if projectsChanged(c.doc.Projects, doc.Projects) || viewsChanged(c.doc.Views, doc.Views) {
		if c.clearCaches != nil {
			c.clearCaches()
		}
	}
	c.doc = doc
	c.mod = mod
	c.poll = doc.EffectivePollSeconds()
	return c.doc, c.poll
}

// writable reports whether the live config can be saved to disk.
func (c *configLive) writable() bool {
	return c != nil && strings.TrimSpace(c.path) != ""
}

// replace saves doc to the config path and updates the in-memory snapshot.
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
	// Re-load so in-memory matches normalized on-disk form.
	loaded, err := config.Load(c.path)
	if err != nil {
		return fmt.Errorf("configlive.replace reload: %w", err)
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return fmt.Errorf("configlive.replace stat: %w", err)
	}
	if projectsChanged(prev.Projects, loaded.Projects) || viewsChanged(prev.Views, loaded.Views) {
		if c.clearCaches != nil {
			c.clearCaches()
		}
	}
	c.doc = loaded
	c.mod = st.ModTime()
	c.poll = loaded.EffectivePollSeconds()
	return nil
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

func projectFingerprint(p config.Project) string {
	return string(p.Host) + "\x00" + p.Path + "\x00" + p.ID + "\x00" + p.LocalPath
}

func viewFingerprint(v config.View) string {
	return v.ID + "\x00" + v.Label + "\x00" + strings.Join(v.Projects, "\x01")
}
