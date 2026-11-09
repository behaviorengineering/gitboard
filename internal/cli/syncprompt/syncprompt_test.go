package syncprompt_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/cli/syncprompt"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

// PromptSyncSources stays GitHub + GitLab only; Azure/Bitbucket use config/API.
func TestPromptSyncSourcesSurfaceIsGitHubGitLabOnly(t *testing.T) {
	titles := promptTitles(t, findFunc(t, "PromptSyncSources"))
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

// findFunc parses syncprompt.go and returns the declaration of name.
func findFunc(t *testing.T, name string) *ast.FuncDecl {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	srcPath := filepath.Join(filepath.Dir(thisFile), "syncprompt.go")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), srcPath, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name != nil && fd.Name.Name == name {
			return fd
		}
	}
	t.Fatalf("%s not found", name)
	return nil
}

// promptTitles collects the string literal passed to every .Title(...) call
// in fn.
func promptTitles(t *testing.T, fn *ast.FuncDecl) []string {
	t.Helper()
	var titles []string
	ast.Inspect(fn, func(n ast.Node) bool {
		if title, ok := titleLiteral(n); ok {
			titles = append(titles, title)
		}
		return true
	})
	return titles
}

// titleLiteral returns the literal of a one-argument .Title("...") call.
func titleLiteral(n ast.Node) (string, bool) {
	call, ok := n.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return "", false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Title" {
		return "", false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	return strings.Trim(lit.Value, `"`), true
}

func TestFormatCandidates(t *testing.T) {
	var buf bytes.Buffer
	syncprompt.FormatCandidates(&buf, []projectsync.Candidate{
		{Index: 1, Tracked: true, Host: boardconfig.HostGitHub, Path: "a/b"},
	})
	if !strings.Contains(buf.String(), "[1] [x] github a/b") {
		t.Fatalf("got %q", buf.String())
	}
}
