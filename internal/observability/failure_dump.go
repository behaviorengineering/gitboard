package observability

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type dumpedTrace struct {
	TraceID           string       `json:"trace_id"`
	DumpedAt          time.Time    `json:"dumped_at"`
	Reason            string       `json:"reason"`
	RootName          string       `json:"root_name,omitempty"`
	RootStatusMessage string       `json:"root_status_message,omitempty"`
	SpanCount         int          `json:"span_count"`
	Spans             []dumpedSpan `json:"spans"`
}

type dumpedSpan struct {
	TraceID       string         `json:"trace_id"`
	SpanID        string         `json:"span_id"`
	ParentSpanID  string         `json:"parent_span_id,omitempty"`
	Name          string         `json:"name"`
	Kind          string         `json:"kind,omitempty"`
	StatusCode    string         `json:"status_code"`
	StatusMessage string         `json:"status_message,omitempty"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	Attributes    map[string]any `json:"attributes,omitempty"`
	Events        []dumpedEvent  `json:"events,omitempty"`
}

type dumpedEvent struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type traceDumpBuffer struct {
	spans    []dumpedSpan
	hasError bool
}

type failureDumpProcessor struct {
	dir         string
	maxAgeHours int
	maxFiles    int
	mu          sync.Mutex
	traces      map[string]*traceDumpBuffer
}

func newFailureDumpProcessor(dir string, maxAgeHours, maxFiles int) *failureDumpProcessor {
	return &failureDumpProcessor{
		dir:         dir,
		maxAgeHours: maxAgeHours,
		maxFiles:    maxFiles,
		traces:      make(map[string]*traceDumpBuffer),
	}
}

func (p *failureDumpProcessor) OnStart(context.Context, sdktrace.ReadWriteSpan) {}

func (p *failureDumpProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p == nil || s == nil {
		return
	}
	snap := snapshotSpan(s)
	traceID := snap.TraceID
	if traceID == "" {
		return
	}
	isRoot := snap.ParentSpanID == ""
	isError := s.Status().Code == codes.Error

	p.mu.Lock()
	defer p.mu.Unlock()
	buf := p.traces[traceID]
	if buf == nil {
		buf = &traceDumpBuffer{}
		p.traces[traceID] = buf
	}
	buf.spans = append(buf.spans, snap)
	if isError {
		buf.hasError = true
	}
	if !isRoot {
		return
	}
	delete(p.traces, traceID)
	if !isError {
		return
	}
	p.writeLocked(dumpedTrace{
		TraceID:           traceID,
		DumpedAt:          time.Now().UTC(),
		Reason:            "root_span_error",
		RootName:          snap.Name,
		RootStatusMessage: snap.StatusMessage,
		Spans:             buf.spans,
	})
}

func (p *failureDumpProcessor) Shutdown(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for tid, buf := range p.traces {
		if buf != nil && buf.hasError {
			p.writeLocked(dumpedTrace{
				TraceID:  tid,
				DumpedAt: time.Now().UTC(),
				Reason:   "shutdown_with_error_spans",
				Spans:    buf.spans,
			})
		}
		delete(p.traces, tid)
	}
	return nil
}

func (p *failureDumpProcessor) ForceFlush(context.Context) error { return nil }

func (p *failureDumpProcessor) writeLocked(doc dumpedTrace) {
	sort.Slice(doc.Spans, func(i, j int) bool {
		return doc.Spans[i].StartTime.Before(doc.Spans[j].StartTime)
	})
	doc.SpanCount = len(doc.Spans)
	if err := os.MkdirAll(p.dir, 0o750); err != nil {
		logf("mkdir dump dir: %v", err)
		return
	}
	path := filepath.Join(p.dir, doc.TraceID+".json")
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		logf("marshal dump: %v", err)
		return
	}
	if err := os.WriteFile(path, body, 0o640); err != nil {
		logf("write dump: %v", err)
		return
	}
	logf("Inference failure dump written path=%s trace_id=%s spans=%d", path, doc.TraceID, doc.SpanCount)
	if err := pruneFailureDumps(p.dir, p.maxAgeHours, p.maxFiles); err != nil {
		logf("prune dumps: %v", err)
	}
}

func pruneFailureDumps(dir string, maxAgeHours, maxFiles int) error {
	if maxAgeHours <= 0 && maxFiles <= 0 {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	type dumpFile struct {
		path    string
		modTime time.Time
	}
	var kept []dumpFile
	now := time.Now()
	maxAge := time.Duration(maxAgeHours) * time.Hour
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		path := filepath.Join(dir, entry.Name())
		if maxAgeHours > 0 && now.Sub(info.ModTime()) > maxAge {
			_ = os.Remove(path)
			continue
		}
		kept = append(kept, dumpFile{path: path, modTime: info.ModTime()})
	}
	if maxFiles <= 0 || len(kept) <= maxFiles {
		return nil
	}
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].modTime.After(kept[j].modTime)
	})
	for _, file := range kept[maxFiles:] {
		_ = os.Remove(file.path)
	}
	return nil
}

func snapshotSpan(s sdktrace.ReadOnlySpan) dumpedSpan {
	sc := s.SpanContext()
	parentID := ""
	if s.Parent().IsValid() {
		parentID = s.Parent().SpanID().String()
	}
	status := s.Status()
	return dumpedSpan{
		TraceID:       sc.TraceID().String(),
		SpanID:        sc.SpanID().String(),
		ParentSpanID:  parentID,
		Name:          s.Name(),
		Kind:          s.SpanKind().String(),
		StatusCode:    statusCodeString(status.Code),
		StatusMessage: redactURLsInText(status.Description),
		StartTime:     s.StartTime().UTC(),
		EndTime:       s.EndTime().UTC(),
		Attributes:    attributesToMap(s.Attributes()),
		Events:        snapshotEvents(s.Events()),
	}
}

func snapshotEvents(events []sdktrace.Event) []dumpedEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]dumpedEvent, 0, len(events))
	for _, ev := range events {
		out = append(out, dumpedEvent{
			Name:       ev.Name,
			Timestamp:  ev.Time.UTC(),
			Attributes: attributesToMap(ev.Attributes),
		})
	}
	return out
}

func attributesToMap(attrs []attribute.KeyValue) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		key := string(attr.Key)
		value := attr.Value.AsInterface()
		if str, ok := value.(string); ok {
			value = sanitizeAttrValue(key, str)
		}
		out[key] = value
	}
	return out
}

func statusCodeString(code codes.Code) string {
	switch code {
	case codes.Ok:
		return "OK"
	case codes.Error:
		return "ERROR"
	default:
		return "UNSET"
	}
}

var embeddedHTTPURL = regexp.MustCompile(`https?://[^\s"'<>]+`)

func redactURLsInText(s string) string {
	if s == "" || !strings.Contains(s, "://") {
		return s
	}
	return embeddedHTTPURL.ReplaceAllStringFunc(s, func(raw string) string {
		if i := strings.Index(raw, "?"); i >= 0 {
			return raw[:i] + "?REDACTED"
		}
		return raw
	})
}

func sanitizeAttrValue(key, value string) string {
	k := strings.ToLower(key)
	if strings.Contains(k, "authorization") || strings.Contains(k, "api_key") || strings.Contains(k, "token") {
		return "REDACTED"
	}
	return redactURLsInText(value)
}

var _ sdktrace.SpanProcessor = (*failureDumpProcessor)(nil)
