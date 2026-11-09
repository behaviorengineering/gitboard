package remotegit

import (
	"context"
	"errors"

	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// ErrClientMissing is returned when a forge client method is called on a nil
// receiver or on a client built without its required dependencies.
var ErrClientMissing = errors.New("forge client missing")

// ErrNilLoader is returned when a cache read has no loader to fall back on.
var ErrNilLoader = errors.New("remotegit: nil loader")

func clientMissing(op, forge string) error {
	return gberr.Wrap(ErrClientMissing, gberr.CodeInternal, op, forge+" client missing")
}

// forgeErr wraps a failed forge read as gberr.CodeUnavailable, or
// gberr.CodeCanceled when the caller cancelled. A deadline stays Unavailable:
// the forge was too slow for the budget. msg is client-safe; the cause (CLI
// stderr, API status) stays in the chain for logs. repo names the project,
// org, group or workspace the read was for.
func forgeErr(err error, op, msg, repo string) *gberr.Error {
	code := gberr.CodeUnavailable
	if errors.Is(err, context.Canceled) {
		code = gberr.CodeCanceled
	}
	return gberr.Wrap(err, code, op, msg).With("repo", repo)
}

// repoPathErr reports a project path the forge cannot address.
func repoPathErr(err error, op, repo string) error {
	return gberr.Wrap(err, gberr.CodeInvalid, op, "project path is not valid for this forge").With("repo", repo)
}
