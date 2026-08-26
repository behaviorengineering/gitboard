package server

import (
	"os"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
)

// configLive reloads config from disk when the file mtime advances.
// When the tracked project set changes, it clears the forge TTL cache
// so the board reflects gitboard sync without a serve restart.
type configLive struct {
	mu    sync.Mutex
	path  string
	mod   time.Time
	doc   config.File
	poll  int
	cache *forge.TTLCache
}

func newConfigLive(path string, doc config.File, poll int, cache *forge.TTLCache) *configLive {
	cl := &configLive{
		path:  path,
		doc:   doc,
		poll:  poll,
		cache: cache,
	}
	if path != "" {
		if st, err := os.Stat(path); err == nil {
			cl.mod = st.ModTime()
		}
	}
	return cl
}

// snapshot returns the current config. When ConfigPath is set and the file
// mtime is newer, it reloads from disk. On reload failure it keeps the last
// good snapshot. Project-set changes clear the forge cache.
func (c *configLive) snapshot() (config.File, int) {
	if c == nil {
		return config.File{}, config.DefaultPollSeconds
	}
	c.mu.Lock()
	defer c.mu.Unlock()
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
		return c.doc, c.poll
	}
	if projectsChanged(c.doc.Projects, doc.Projects) {
		c.cache.Clear()
	}
	c.doc = doc
	c.mod = mod
	c.poll = doc.EffectivePollSeconds()
	return c.doc, c.poll
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

func projectFingerprint(p config.Project) string {
	return string(p.Host) + "\x00" + p.Path + "\x00" + p.ID + "\x00" + p.LocalPath
}
