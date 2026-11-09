package dashboard

import (
	"context"
	"strings"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// FailedJobs lists the failed jobs of one CI run of a tracked project.
// An unknown project is CodeInvalid; a forge failure is CodeUnavailable.
func (s *Service) FailedJobs(ctx context.Context, doc boardconfig.File, projectID, runID string) ([]board.FailedJob, error) {
	const op = "dashboard.FailedJobs"
	p, client, err := s.forgeFor(op, doc, projectID)
	if err != nil {
		return nil, err
	}
	jobs, err := client.FailedJobs(ctx, p, strings.TrimSpace(runID))
	if err != nil {
		return nil, ciErr(err, op, "failed to list jobs").With("project_id", p.ID).With("run_id", runID)
	}
	return jobs, nil
}

// JobLog returns the tail of one CI job's log for a tracked project.
// An unknown project is CodeInvalid; a forge failure is CodeUnavailable.
func (s *Service) JobLog(ctx context.Context, doc boardconfig.File, projectID, runID, jobID string) (string, error) {
	const op = "dashboard.JobLog"
	p, client, err := s.forgeFor(op, doc, projectID)
	if err != nil {
		return "", err
	}
	logText, err := client.JobLog(ctx, p, runID, jobID)
	if err != nil {
		return "", ciErr(err, op, "failed to fetch job log").
			With("project_id", p.ID).With("run_id", runID).With("job_id", jobID)
	}
	return logText, nil
}

// forgeFor returns the tracked project with id and its forge client.
func (s *Service) forgeFor(op string, doc boardconfig.File, id string) (boardconfig.Project, remotegit.Client, error) {
	if s == nil {
		return boardconfig.Project{}, nil, unavailable(op, ErrServiceMissing)
	}
	p, ok := FindProject(doc.Projects, strings.TrimSpace(id))
	if !ok {
		return boardconfig.Project{}, nil, invalid(op, "unknown project")
	}
	client := s.clientFor(p.Host)
	if client == nil {
		return boardconfig.Project{}, nil, gberr.Wrap(remotegit.ErrClientMissing, gberr.CodeInternal, op, "forge client missing").
			With("host", string(p.Host))
	}
	return p, client, nil
}

// ciErr keeps a forge's own CodeInvalid (bad run or job id, or a forge that
// has no job logs yet) and reports everything else as an unavailable forge
// with msg.
func ciErr(err error, op, msg string) *gberr.Error {
	if e, ok := gberr.As(err); ok && e.Code == gberr.CodeInvalid {
		return gberr.Wrap(err, gberr.CodeInvalid, op, e.Msg)
	}
	return gberr.Wrap(err, gberr.CodeUnavailable, op, msg)
}
