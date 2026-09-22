package dashboard

import (
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Shell builds a config-only dashboard for a view: project stubs with identity
// fields and empty branches. No forge or local git calls.
func Shell(doc config.File, viewID string) (board.Dashboard, error) {
	out := board.Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Views:       ViewSummaries(doc),
	}
	projects, view, err := doc.ProjectsForView(viewID)
	if err != nil {
		return out, err
	}
	out.ActiveView = view.ID
	rows := make([]board.ProjectSummary, len(projects))
	for i, p := range projects {
		rows[i] = board.ProjectSummary{
			ID:           p.ID,
			Label:        p.Label,
			Host:         string(p.Host),
			Path:         p.Path,
			Org:          forgeOrg(p.Path),
			OpenURL:      p.OpenURL(),
			Capabilities: remotegit.HostCapabilities(p.Host),
		}
	}
	out.Projects = rows
	return out, nil
}
