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

// TestInspectorHasNoMutators keeps Inspector read-only: any method that looks
// like a git writer belongs on Mutator, which only a held MutationLease hands out.
func TestInspectorHasNoMutators(t *testing.T) {
	for name := range methodsOf(t, "Inspector") {
		if looksLikeMutator(name) {
			t.Errorf("Inspector.%s looks like a git writer; move it onto Mutator so it runs under the lease", name)
		}
	}
}

// TestMutatorExportsBindToLease checks that every exported Mutator method calls
// bind, so a path outside the lease's common git directory is refused.
func TestMutatorExportsBindToLease(t *testing.T) {
	methods := methodsOf(t, "Mutator")
	if len(methods) == 0 {
		t.Fatal("no Mutator methods found")
	}
	for name, fn := range methods {
		if !ast.IsExported(name) {
			continue
		}
		if !callsSelector(fn, "bind") {
			t.Errorf("Mutator.%s must call m.bind before running git", name)
		}
	}
}

func methodsOf(t *testing.T, recvName string) map[string]*ast.FuncDecl {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	found := map[string]*ast.FuncDecl{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range src.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name == nil {
				continue
			}
			if receiverName(fn.Recv) == recvName {
				found[fn.Name.Name] = fn
			}
		}
	}
	return found
}

func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	switch t := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func callsSelector(fn *ast.FuncDecl, sel string) bool {
	found := false
	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if s, ok := call.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == sel {
			found = true
			return false
		}
		return true
	})
	return found
}

func looksLikeMutator(name string) bool {
	lower := strings.ToLower(name)
	prefixes := []string{"fetch", "pull", "push", "remove", "reset", "ensure", "update", "write", "delete", "switch", "checkout", "merge"}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}
