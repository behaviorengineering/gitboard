package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// errorCodeHeader carries the gberr.Code of a failed request so clients can
// branch on a stable value instead of the message text.
const errorCodeHeader = "X-Gitboard-Error-Code"

// statusFor maps a domain error code to an HTTP status. It is the one place
// the server turns gberr codes into statuses.
func statusFor(code gberr.Code) int {
	switch code {
	case gberr.CodeInvalid:
		return http.StatusBadRequest
	case gberr.CodeNotFound:
		return http.StatusNotFound
	case gberr.CodeConflict, gberr.CodeConfirmRequired:
		return http.StatusConflict
	case gberr.CodeCanceled:
		return http.StatusGatewayTimeout
	case gberr.CodeUnavailable:
		return http.StatusServiceUnavailable
	case gberr.CodeInternal:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

// publicMessage is the only error text a client receives: the domain error's
// client-safe Msg, or a fixed text for the code. Raw err.Error() never leaves
// the server because it carries local paths, CLI stderr and config detail.
func publicMessage(err error) string {
	code := gberr.CodeOf(err)
	if e, ok := gberr.As(err); ok && e.Msg != "" && code != gberr.CodeInternal {
		return e.Msg
	}
	switch code {
	case gberr.CodeInvalid:
		return "invalid request"
	case gberr.CodeNotFound:
		return "not found"
	case gberr.CodeConflict:
		return "conflict with current state"
	case gberr.CodeConfirmRequired:
		return "confirmation required"
	case gberr.CodeCanceled:
		return "request cancelled or timed out"
	case gberr.CodeUnavailable:
		return "a dependency is unavailable"
	case gberr.CodeInternal:
		return "internal error"
	}
	return "internal error"
}

// logError records the full cause chain with discriminator fields. Server
// faults log at error level; client and state outcomes at info level.
func logError(log observability.Logger, op string, err error) {
	code := gberr.CodeOf(err)
	fields := []any{"op", op, "code", string(code)}
	extra := gberr.FieldsOf(err)
	keys := make([]string, 0, len(extra))
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fields = append(fields, k, extra[k])
	}
	l := log.With(fields...)
	if statusFor(code) >= http.StatusInternalServerError {
		l.Errorf("request failed: %v", err)
		return
	}
	l.Infof("request refused: %v", err)
}

// writeError logs err and sends the client a stable message, the status for
// its code and the code in errorCodeHeader. A confirm-required error sends the
// dashboard.ConfirmRequiredError JSON the UI needs to ask and retry.
func writeError(w http.ResponseWriter, log observability.Logger, op string, err error) {
	if err == nil {
		return
	}
	logError(log, op, err)
	code := gberr.CodeOf(err)
	w.Header().Set(errorCodeHeader, string(code))
	if code == gberr.CodeConfirmRequired {
		var cre dashboard.ConfirmRequiredError
		if errors.As(err, &cre) {
			writeJSONStatus(w, log, statusFor(code), cre)
			return
		}
	}
	http.Error(w, publicMessage(err), statusFor(code))
}

// writeJSON sends v as a 200 JSON response.
func writeJSON(w http.ResponseWriter, log observability.Logger, v any) {
	writeJSONStatus(w, log, http.StatusOK, v)
}

// writeJSONStatus sets the JSON headers before the status line, then encodes
// v. Once the status is sent, logging is the only useful action left when
// encoding fails.
func writeJSONStatus(w http.ResponseWriter, log observability.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.With("status", status).Errorf("encode response: %v", err)
	}
}
