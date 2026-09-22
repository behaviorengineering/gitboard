package observability

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	var b strings.Builder
	b.WriteString(l.prefix)
	for i := 0; i+1 < len(fields); i += 2 {
		_, _ = fmt.Fprintf(&b, "%v=%v ", fields[i], fields[i+1])
	}
	return &stdLogger{out: l.out, prefix: b.String()}
}
