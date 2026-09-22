package dashboard

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// InvalidateProject drops forge TTL state for one tracked project.
func (s *Service) InvalidateProject(host, path string) {
	if s == nil || s.Cache == nil {
		return
	}
	s.Cache.ForgetProject(remotegit.CacheKey(host, strings.Trim(path, "/")))
}

// InvalidateOrigin drops the origin-fetch TTL for one common git dir.
func (s *Service) InvalidateOrigin(commonDir string) {
	if s == nil || s.OriginFetch == nil {
		return
	}
	s.OriginFetch.Invalidate(commonDir)
}

// mappedPathAllowlist returns whether abs belongs to the project without a full
// attachLocal (no origin fetch, no EnrichOriginSync). Uses cached scan when possible.
func (s *Service) mappedPathAllowlist(ctx context.Context, doc config.File, p config.Project, abs string) (bool, error) {
	if s == nil || s.Local == nil {
		return false, ErrLocalInspectorMissing
	}
	abs = filepath.Clean(abs)
	explicit := strings.TrimSpace(p.LocalPath)
	if explicit != "" {
		exp, err := localgit.ExpandPath(explicit)
		if err == nil {
			if resolved, err := filepath.EvalSymlinks(exp); err == nil {
				exp = resolved
			}
			if filepath.Clean(exp) == abs {
				return true, nil
			}
		}
	}

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		scanTTL := time.Duration(doc.EffectiveScanSeconds()) * time.Second
		disc = s.scanRootsCached(ctx, doc.Local.Roots, scanTTL, false)
	}
	checkouts := disc.Appearances(string(p.Host), p.Path)
	if explicit != "" {
		if exp, err := localgit.ExpandPath(explicit); err == nil {
			c := localgit.Checkout{Path: exp, Role: localgit.RoleStandalone}
			localgit.FillCheckoutMeta(&c)
			checkouts = append(checkouts, c)
		}
	}
	local := &board.LocalStatus{Mapped: len(checkouts) > 0}
	for _, c := range checkouts {
		local.Appearances = append(local.Appearances, board.LocalAppearance{Path: c.Path})
		if local.Path == "" {
			local.Path = c.Path
		}
	}
	if !local.Mapped && explicit == "" {
		return false, nil
	}
	if local.Path == "" && explicit != "" {
		if exp, err := localgit.ExpandPath(explicit); err == nil {
			local.Path = exp
			local.Mapped = true
		}
	}
	return repoPathAllowed(local, abs), nil
}
