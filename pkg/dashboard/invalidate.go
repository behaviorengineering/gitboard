package dashboard

import (
	"context"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// InvalidateProject drops forge TTL state for one tracked project. The key is
// remotegit.CacheKey, the one the summary loader writes.
func (s *Service) InvalidateProject(host boardconfig.Host, path string) {
	if s == nil {
		return
	}
	s.cache.ForgetProject(remotegit.CacheKey(host, path))
}

// InvalidateOrigin drops the origin-fetch TTL for one common git dir.
func (s *Service) InvalidateOrigin(commonDir string) {
	if s == nil {
		return
	}
	s.originFetch.Invalidate(commonDir)
}

// mappedPathAllowlist returns whether abs belongs to the project without a full
// attachLocal (no origin fetch, no origin sync). Uses cached scan when possible.
func (s *Service) mappedPathAllowlist(ctx context.Context, doc boardconfig.File, p boardconfig.Project, abs string) (bool, error) {
	if s == nil || s.local == nil {
		return false, unavailable("dashboard.mappedPathAllowlist", ErrLocalInspectorMissing)
	}
	abs = cleanRepoPath(abs)
	explicit := explicitLocalPath(p)
	if explicit != "" && cleanRepoPath(explicit) == abs {
		return true, nil
	}

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		scanTTL := time.Duration(doc.EffectiveScanSeconds()) * time.Second
		disc = s.scanRootsCached(ctx, doc.Local.Roots, scanTTL, false)
	}
	checkouts := disc.Appearances(p.Host, p.Path)
	if explicit != "" {
		c := localgit.Checkout{Path: explicit, Role: board.AppearanceStandalone}
		localgit.FillCheckoutMeta(&c)
		checkouts = append(checkouts, c)
	}
	if len(checkouts) == 0 {
		return false, nil
	}
	local := &board.LocalStatus{Mapped: true, Path: checkouts[0].Path}
	for _, c := range checkouts {
		local.Appearances = append(local.Appearances, board.LocalAppearance{Path: c.Path})
	}
	return repoPathAllowed(local, abs), nil
}

// explicitLocalPath is p.LocalPath expanded, or "" when unset or unusable.
func explicitLocalPath(p boardconfig.Project) string {
	raw := strings.TrimSpace(p.LocalPath)
	if raw == "" {
		return ""
	}
	exp, err := localgit.ExpandPath(raw)
	if err != nil {
		return ""
	}
	return exp
}
