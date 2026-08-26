package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func TestFailureDumpOnRootError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tp, err := Init(InitConfig{
		ServiceName:            "gitboard-test",
		FailureDumpDir:         dir,
		FailureDumpMaxAgeHours: 48,
		FailureDumpMaxFiles:    20,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Shutdown(t.Context(), tp) })

	tr := otel.Tracer("test")
	_, span := tr.Start(t.Context(), "investigate")
	span.SetStatus(codes.Error, "boom")
	span.End()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e.Name())
		}
	}
	if len(jsonFiles) != 1 {
		t.Fatalf("want 1 dump in %s, got %v", dir, jsonFiles)
	}
	data, err := os.ReadFile(filepath.Join(dir, jsonFiles[0]))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "root_span_error") {
		t.Fatalf("dump missing reason: %s", data)
	}
}
