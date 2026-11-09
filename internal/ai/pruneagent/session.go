package pruneagent

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/behaviorengineering/strop/agentsession"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// ErrSessionNotFound means no stored session or session file has the
// requested id. Session and FailurePath wrap it in gberr.CodeNotFound.
var ErrSessionNotFound = errors.New("pruneagent: session not found")

// errServiceMissing is returned by methods on a nil Service or one built
// without CreateService.
var errServiceMissing = errors.New("pruneagent: service missing")

// IsSessionNotFound reports whether err means the session (or the requested
// session file) does not exist.
func IsSessionNotFound(err error) bool {
	return errors.Is(err, ErrSessionNotFound)
}

// SessionView is one stored investigation for the session read API.
type SessionView struct {
	Meta       *agentsession.Meta  `json:"meta"`
	Card       Card                `json:"card"`
	Turns      []agentsession.Turn `json:"turns"`
	CardError  string              `json:"card_error,omitempty"`
	TurnsError string              `json:"turns_error,omitempty"`
	// PartialErr joins the card and transcript read failures behind
	// CardError and TurnsError, for the caller to log. Never sent to clients.
	PartialErr error `json:"-"`
}

// ready checks the dependencies every Service method needs.
func (s *Service) ready(op string) error {
	if s == nil || s.Store == nil || s.in == nil || s.mutations == nil || s.clock == nil {
		return gberr.Wrap(errServiceMissing, gberr.CodeInternal, op, "prune agent is not configured")
	}
	return nil
}

// storeErr wraps a session store failure as gberr.CodeUnavailable.
func storeErr(err error, op, action, sessionID string) error {
	e := gberr.Wrap(err, gberr.CodeUnavailable, op, "session store failed: "+action)
	if sessionID != "" {
		e = e.With("session_id", sessionID)
	}
	return e
}

// Session loads one stored investigation. A missing session is
// gberr.CodeNotFound (IsSessionNotFound), a malformed id is
// gberr.CodeInvalid, and any other read failure of the session metadata is
// gberr.CodeUnavailable. A missing card or transcript is normal for a session
// still in progress; other card or transcript failures are reported in
// CardError, TurnsError and PartialErr.
func (s *Service) Session(ctx context.Context, id string) (SessionView, error) {
	const op = "pruneagent.Session"
	if s == nil || s.Store == nil {
		return SessionView{}, gberr.Wrap(errServiceMissing, gberr.CodeInternal, op, "prune agent is not configured")
	}
	if _, err := s.Store.Dir(id); err != nil {
		return SessionView{}, gberr.Wrap(err, gberr.CodeInvalid, op, "invalid session id")
	}
	meta, err := s.Store.Load(ctx, id)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return SessionView{}, gberr.Wrap(fmt.Errorf("%w: %w", ErrSessionNotFound, err), gberr.CodeNotFound, op, "session not found").
				With("session_id", id)
		}
		return SessionView{}, storeErr(err, op, "load session", id)
	}
	view := SessionView{Meta: meta}
	var partial []error
	if err := s.Store.LoadJSON(ctx, id, agentsession.FileCard, &view.Card); err != nil && !errors.Is(err, fs.ErrNotExist) {
		view.CardError = "failed to load card"
		partial = append(partial, fmt.Errorf("load card: %w", err))
	}
	turns, err := s.Store.ReadTurns(ctx, id)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		view.TurnsError = "failed to load turns"
		partial = append(partial, fmt.Errorf("read turns: %w", err))
	}
	view.Turns = turns
	view.PartialErr = errors.Join(partial...)
	return view, nil
}

// FailurePath returns the on-disk path of a session's LLM failure dump.
// A malformed id is gberr.CodeInvalid; a missing dump is gberr.CodeNotFound.
func (s *Service) FailurePath(id string) (string, error) {
	const op = "pruneagent.FailurePath"
	if s == nil || s.Store == nil {
		return "", gberr.Wrap(errServiceMissing, gberr.CodeInternal, op, "prune agent is not configured")
	}
	dir, err := s.Store.Dir(id)
	if err != nil {
		return "", gberr.Wrap(err, gberr.CodeInvalid, op, "invalid session id")
	}
	path := filepath.Join(dir, agentsession.FileFailure)
	st, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", gberr.Wrap(fmt.Errorf("%w: %w", ErrSessionNotFound, err), gberr.CodeNotFound, op, "failure dump not found").
				With("session_id", id)
		}
		return "", storeErr(err, op, "stat failure dump", id)
	}
	if st.IsDir() {
		return "", gberr.Wrap(ErrSessionNotFound, gberr.CodeNotFound, op, "failure dump not found").With("session_id", id)
	}
	return path, nil
}
