package syncproj_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// PromptSyncSources stays GitHub + GitLab only; Azure/Bitbucket use config/API.
func TestPromptSyncSourcesSurfaceIsGitHubGitLabOnly(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	srcPath := filepath.Join(filepath.Dir(thisFile), "select.go")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, srcPath, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var fn *ast.FuncDecl
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil || fd.Name.Name != "PromptSyncSources" {
			continue
		}
		fn = fd
		break
	}
	if fn == nil {
		t.Fatal("PromptSyncSources not found")
	}
	var titles []string
	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Title" {
			return true
		}
		if len(call.Args) != 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			return true
		}
		titles = append(titles, strings.Trim(lit.Value, `"`))
		return true
	})
	joined := strings.ToLower(strings.Join(titles, " "))
	if !strings.Contains(joined, "github") || !strings.Contains(joined, "gitlab") {
		t.Fatalf("want GitHub and GitLab titles, got %v", titles)
	}
	for _, bad := range []string{"azure", "bitbucket", "devops"} {
		if strings.Contains(joined, bad) {
			t.Fatalf("interactive sync prompt must not advertise %s: %v", bad, titles)
		}
	}
}
