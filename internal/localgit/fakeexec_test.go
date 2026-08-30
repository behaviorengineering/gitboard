package localgit

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// fakeExec returns canned git output matched by argv substrings (longest match wins).
type fakeExec struct {
	mu        sync.Mutex
	responses map[string][]byte
	errors    map[string]error
	calls     []string
}

func (f *fakeExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *fakeExec) match(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, key)
	var best string
	for substr := range f.responses {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	for substr := range f.errors {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	if err, ok := f.errors[best]; ok {
		return nil, err
	}
	return f.responses[best], nil
}

func (f *fakeExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

func (f *fakeExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f.Run(ctx, name, args...)
}
