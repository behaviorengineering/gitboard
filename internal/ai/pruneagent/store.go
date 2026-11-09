package pruneagent

import (
	"context"

	"github.com/behaviorengineering/strop/agentsession"
)

// SessionWriter records one investigation: the session, its JSON files and turns.
type SessionWriter interface {
	Create(ctx context.Context, kind string, extra map[string]any) (*agentsession.Meta, error)
	SaveJSON(ctx context.Context, id, name string, v any) error
	AppendTurn(ctx context.Context, id string, turn agentsession.Turn) error
}

// SessionReader serves stored sessions to the session read APIs.
type SessionReader interface {
	Dir(id string) (string, error)
	Load(ctx context.Context, id string) (*agentsession.Meta, error)
	LoadJSON(ctx context.Context, id, name string, dest any) error
	ReadTurns(ctx context.Context, id string) ([]agentsession.Turn, error)
}

// SessionStore is the agentsession persistence a Service owns: Investigate
// writes through it and the session APIs read from it.
type SessionStore interface {
	SessionWriter
	SessionReader
}
