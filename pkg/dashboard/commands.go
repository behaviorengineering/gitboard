package dashboard

import (
	"errors"

	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// Commands wraps Service for mutating operations: prune, pull, and sync/views config.
// Read operations (Collect, ClearCaches) remain on Service.
type Commands struct {
	*Service
}

// Sentinel errors for missing command wiring.
var (
	ErrCommandsUnavailable   = errors.New("sync commands unavailable")
	ErrLocalInspectorMissing = errors.New("local git inspector missing")
)

// IsMutationCanceled reports whether err means a mutation lease wait was canceled
// (request context ended while queued for the per-repository coordinator).
func IsMutationCanceled(err error) bool {
	return errors.Is(err, localgit.ErrMutationCanceled)
}

// NewCommands returns a Commands backed by the given Service.
// Panics if s is nil (required dependency at wire-up).
func NewCommands(s *Service) *Commands {
	if s == nil {
		panic("dashboard.NewCommands: Service is required")
	}
	return &Commands{Service: s}
}

func (c *Commands) requireService() error {
	if c == nil || c.Service == nil {
		return ErrCommandsUnavailable
	}
	return nil
}

func (c *Commands) requireLocal() (*Service, error) {
	if err := c.requireService(); err != nil {
		return nil, err
	}
	if c.Local == nil {
		return nil, ErrLocalInspectorMissing
	}
	return c.Service, nil
}
