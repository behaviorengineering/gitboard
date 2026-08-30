package pruneagent

import (
	"context"

	"github.com/behaviorengineering/strop/agentsession"
)

// SessionStore is the agentsession persistence surface used by Investigate and session read APIs.
type SessionStore interface {
	Create(ctx context.Context, kind string, extra map[string]any) (*agentsession.Meta, error)
	SaveJSON(ctx context.Context, id, name string, v any) error
	AppendTurn(ctx context.Context, id string, turn agentsession.Turn) error
	Dir(id string) (string, error)
	Load(ctx context.Context, id string) (*agentsession.Meta, error)
	LoadJSON(ctx context.Context, id, name string, dest any) error
	ReadTurns(ctx context.Context, id string) ([]agentsession.Turn, error)
}
