package dashboard

import (
	"errors"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// Commands wraps Service for mutating operations: prune, pull, and sync/views config.
// Read operations (Collect, ClearCaches) remain on Service.
type Commands struct {
	*Service
}

// Sentinel errors for missing service wiring. Commands wrap them in a
// gberr.Error with gberr.CodeUnavailable; errors.Is still matches.
var (
	ErrServiceMissing        = errors.New("dashboard service missing")
	ErrCommandsUnavailable   = errors.New("sync commands unavailable")
	ErrLocalInspectorMissing = errors.New("local git inspector missing")
)

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
		return unavailable("dashboard.Commands", ErrCommandsUnavailable)
	}
	return nil
}

func (c *Commands) requireLocal() (*Service, error) {
	if err := c.requireService(); err != nil {
		return nil, err
	}
	if c.local == nil {
		return nil, unavailable("dashboard.Commands", ErrLocalInspectorMissing)
	}
	return c.Service, nil
}

// unavailable wraps a missing-wiring sentinel; its text is client-safe.
func unavailable(op string, sentinel error) error {
	return gberr.Wrap(sentinel, gberr.CodeUnavailable, op, sentinel.Error())
}
