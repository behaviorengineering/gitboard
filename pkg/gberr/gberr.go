// Package gberr is public API: gitboard's typed domain error. Library
// packages (dashboard, remotegit, localgit) return *Error at their boundary so
// callers branch on a stable Code instead of message text, and entry packages
// (the HTTP server, the CLI) map each Code to a status or exit in one place.
//
// The package does not know about HTTP or process exit codes.
package gberr

import (
	"context"
	"errors"
	"maps"
)

// Code is a stable, machine-readable error class.
type Code string

// Stable codes. Callers MUST pass these constants, never string literals.
const (
	// CodeInvalid means the caller's input is wrong (bad field, unknown id).
	CodeInvalid Code = "invalid"
	// CodeNotFound means the named resource does not exist.
	CodeNotFound Code = "not_found"
	// CodeConflict means current state refuses the action (dirty tree,
	// diverged branch, branch no longer safe to remove, lock held).
	CodeConflict Code = "conflict"
	// CodeConfirmRequired means the operator must approve a follow-up and
	// retry (for example clearing a stale index.lock).
	CodeConfirmRequired Code = "confirm_required"
	// CodeCanceled means the caller's context ended or a wait was cancelled.
	CodeCanceled Code = "canceled"
	// CodeUnavailable means a dependency failed: a forge CLI or API, git, the
	// LLM endpoint or the session store.
	CodeUnavailable Code = "unavailable"
	// CodeInternal means a gitboard defect or an unclassified failure.
	CodeInternal Code = "internal"
)

// Error is a typed domain error with a cause chain.
//
// Msg is the client-safe text: it MUST NOT carry local paths, CLI stderr or
// other cause text, because entry packages may send it to an HTTP client.
// Op, Fields and Cause are for logs only.
type Error struct {
	Code   Code
	Op     string
	Msg    string
	Fields map[string]string
	Cause  error
}

// New builds a domain error with no cause.
func New(code Code, op, msg string) *Error {
	return &Error{Code: code, Op: op, Msg: msg}
}

// Wrap attaches a code, operation and client-safe message to cause.
func Wrap(cause error, code Code, op, msg string) *Error {
	return &Error{Code: code, Op: op, Msg: msg, Cause: cause}
}

// With adds a troubleshooting field (logged, never sent to clients) and
// returns e for chaining.
func (e *Error) With(key, value string) *Error {
	if e == nil {
		return nil
	}
	if e.Fields == nil {
		e.Fields = map[string]string{}
	}
	e.Fields[key] = value
	return e
}

// Error returns "op: msg: cause" with empty parts left out.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	out := e.Op
	for _, part := range []string{e.Msg, causeText(e.Cause)} {
		if part == "" {
			continue
		}
		if out != "" {
			out += ": "
		}
		out += part
	}
	if out == "" {
		return string(e.Code)
	}
	return out
}

// Unwrap exposes the cause to errors.Is and errors.As.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// As returns the outermost *Error in err's chain.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) && e != nil {
		return e, true
	}
	return nil, false
}

// CodeOf classifies err. It returns the outermost *Error's Code, CodeCanceled
// for a bare context cancellation or deadline, CodeInternal for any other
// error, and "" for nil.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	if e, ok := As(err); ok && e.Code != "" {
		return e.Code
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return CodeCanceled
	}
	return CodeInternal
}

// HasCode reports whether CodeOf(err) is code.
func HasCode(err error, code Code) bool {
	return err != nil && CodeOf(err) == code
}

// FieldsOf merges Fields from every *Error in err's chain; outer values win.
func FieldsOf(err error) map[string]string {
	out := map[string]string{}
	var chain []*Error
	for cur := err; cur != nil; {
		e, ok := As(cur)
		if !ok {
			break
		}
		chain = append(chain, e)
		cur = e.Cause
	}
	for i := len(chain) - 1; i >= 0; i-- {
		maps.Copy(out, chain[i].Fields)
	}
	return out
}

func causeText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
