package localgit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// knownMutators lists Inspector methods that must acquire MutationCoordinator
// ownership at the dashboard (or equivalent) call boundary. Update this list
// when adding a new local Git writer, and wire the coordinator in the same change.
var knownMutators = map[string]struct{}{
	"FetchOrigin":                    {},
	"FetchOriginBranches":            {},
	"FetchOriginCached":              {},
	"FetchOriginSmart":               {},
	"PullFFOnly":                     {},
	"PullFFOnlyWithPhases":           {},
	"RemoveSafeCheckout":             {},
	"EnsureWritableIndex":            {},
	"EnsureWritableIndexAge":         {},
	"InspectSync":                    {},
	"resetUnusedBranchToOrigin":      {},
	"ensureDefaultReadyForSafePrune": {},
	"ensureBranchFFFromOrigin":       {},
	"updateSubmodules":               {},
}

func TestMutatorInventoryIsDocumented(t *testing.T) {
	root := "."
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	found := map[string]struct{}{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := parser.ParseFile(fset, filepath.Join(root, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range src.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name == nil {
				continue
			}
			if !receiverIsInspector(fn.Recv) {
				continue
			}
			n := fn.Name.Name
			if looksLikeMutator(n, fn) {
				found[n] = struct{}{}
			}
		}
	}

	for name := range found {
		if _, ok := knownMutators[name]; !ok {
			t.Errorf("Inspector.%s looks like a Git mutator but is not in knownMutators; wire MutationCoordinator and update the inventory", name)
		}
	}
	for name := range knownMutators {
		if _, ok := found[name]; !ok {
			t.Errorf("knownMutators lists %s but no Inspector method was found; remove stale inventory entry", name)
		}
	}
}

func receiverIsInspector(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	switch t := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name == "Inspector"
		}
	case *ast.Ident:
		return t.Name == "Inspector"
	}
	return false
}

func looksLikeMutator(name string, fn *ast.FuncDecl) bool {
	lower := strings.ToLower(name)
	prefixes := []string{"fetch", "pull", "push", "remove", "reset", "ensure", "update", "write", "delete", "switch", "checkout", "merge"}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	if strings.Contains(lower, "inspectsync") {
		return true
	}
	_ = fn
	return false
}
