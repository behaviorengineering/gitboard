package servertiming

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Recorder accumulates Server-Timing metrics for one request.
type Recorder struct {
	mu      sync.Mutex
	metrics []metric
}

type metric struct {
	name string
	dur  time.Duration
	desc string
}

// New returns an empty recorder.
func New() *Recorder {
	return &Recorder{}
}

// Track runs fn and records elapsed time under name.
func (r *Recorder) Track(name string, fn func()) {
	if r == nil || fn == nil {
		if fn != nil {
			fn()
		}
		return
	}
	start := time.Now()
	fn()
	r.Add(name, time.Since(start), "")
}

// Add records a timing sample.
func (r *Recorder) Add(name string, d time.Duration, desc string) {
	if r == nil || name == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = append(r.metrics, metric{name: name, dur: d, desc: desc})
}

// Header returns the Server-Timing header value.
func (r *Recorder) Header() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.metrics) == 0 {
		return ""
	}
	parts := make([]string, 0, len(r.metrics))
	for _, m := range r.metrics {
		ms := float64(m.dur) / float64(time.Millisecond)
		part := fmt.Sprintf("%s;dur=%.1f", m.name, ms)
		if m.desc != "" {
			part += fmt.Sprintf(`;desc="%s"`, strings.ReplaceAll(m.desc, `"`, ""))
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

// WriteHeader sets Server-Timing on w when metrics exist.
func (r *Recorder) WriteHeader(w http.ResponseWriter) {
	if r == nil || w == nil {
		return
	}
	if h := r.Header(); h != "" {
		w.Header().Set("Server-Timing", h)
	}
}
