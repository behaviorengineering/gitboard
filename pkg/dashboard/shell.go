package dashboard

import (
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Shell builds a config-only dashboard for a view: project stubs with identity
// fields and empty branches. No forge or local git calls.
func (s *Service) Shell(doc boardconfig.File, viewID string) (board.Dashboard, error) {
	if s == nil {
		return board.Dashboard{}, unavailable("dashboard.Shell", ErrServiceMissing)
	}
	out := board.Dashboard{
		GeneratedAt: s.clock().UTC().Format(time.RFC3339),
		Views:       ViewSummaries(doc),
	}
	projects, view, err := doc.ProjectsForView(viewID)
	if err != nil {
		return out, invalidCause("dashboard.Shell", "unknown view", err)
	}
	out.ActiveView = view.ID
	rows := make([]board.ProjectSummary, len(projects))
	for i, p := range projects {
		rows[i] = remotegit.BaseSummary(p)
	}
	out.Projects = rows
	return out, nil
}
