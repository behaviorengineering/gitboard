// Package observability holds the injectable process logger. Tracing and
// failure dumps come from olly (github.com/behaviorengineering/olly), set up
// at the CLI entrypoints.
package observability

import (
	"fmt"
	"log"
	"os"
)

// Logger is the injectable structured logger for services (C14).
type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	With(fields ...any) Logger
}

type stdLogger struct {
	out    *log.Logger
	prefix string
}

// NewLogger returns a stderr logger with optional key=value fields in the prefix.
func NewLogger() Logger {
	return &stdLogger{out: log.New(os.Stderr, "", 0)}
}

func (l *stdLogger) Infof(format string, args ...any) {
	l.out.Printf(l.prefix+"INFO "+format, args...)
}

func (l *stdLogger) Errorf(format string, args ...any) {
	l.out.Printf(l.prefix+"ERROR "+format, args...)
}

func (l *stdLogger) With(fields ...any) Logger {
	if len(fields) == 0 {
		return l
	}
	prefix := []byte(l.prefix)
	for i := 0; i+1 < len(fields); i += 2 {
		prefix = fmt.Appendf(prefix, "%v=%v ", fields[i], fields[i+1])
	}
	return &stdLogger{out: l.out, prefix: string(prefix)}
}
