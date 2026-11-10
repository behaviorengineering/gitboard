package server

import (
	"strings"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/board"
)

const leakyText = "git -C /home/me/src/repo fetch: fatal: unable to access 'https://***@github.com/o/r/': 403"

func leakyRow() board.ProjectSummary {
	return board.ProjectSummary{
		ID:    "p1",
		Error: leakyText,
		Local: &board.LocalStatus{
			Mapped: true,
			Path:   "/home/me/src/repo",
			Error:  leakyText,
			Appearances: []board.LocalAppearance{
				{Path: "/home/me/src/repo", Error: leakyText},
				{Path: "/home/me/src/other"},
			},
		},
	}
}

func TestDetailRedactorOff(t *testing.T) {
	d := detailRedactor{on: false, log: observability.NewLogger()}
	row := leakyRow()
	d.project(&row)
	if row.Error != leakyText || row.Local.Error != leakyText || row.Local.Appearances[0].Error != leakyText {
		t.Fatalf("loopback must keep error text: %+v", row)
	}
	lines := []string{"skipped /home/me/x"}
	if got := d.warnings("op", lines); len(got) != 1 || got[0] != lines[0] {
		t.Fatalf("loopback must keep warnings, got %q", got)
	}
}

func TestDetailRedactorOn(t *testing.T) {
	d := detailRedactor{on: true, log: observability.NewLogger()}
	row := leakyRow()
	d.project(&row)
	if row.Error != redactedProjectError {
		t.Fatalf("project error = %q", row.Error)
	}
	if row.Local.Error != redactedLocalError || row.Local.Appearances[0].Error != redactedLocalError {
		t.Fatalf("local errors not redacted: %+v", row.Local)
	}
	if row.Local.Appearances[1].Error != "" {
		t.Fatalf("empty error must stay empty, got %q", row.Local.Appearances[1].Error)
	}
	if row.Local.Path != "/home/me/src/repo" {
		t.Fatalf("board paths are content and must stay, got %q", row.Local.Path)
	}

	var tooling board.Tooling
	tooling.GitHub.Authed = true
	tooling.GitHub.Detail = "Logged in to github.com account me (/home/me/.config/gh/hosts.yml)"
	d.tooling(&tooling)
	if tooling.GitHub.Detail != "" || !tooling.GitHub.Authed {
		t.Fatalf("tooling = %+v", tooling.GitHub)
	}

	got := d.warnings("op", []string{"skipped /home/me/a", "skipped /home/me/b"})
	if len(got) != 1 || strings.Contains(got[0], "/home/me") || !strings.HasPrefix(got[0], "2 warnings") {
		t.Fatalf("warnings = %q", got)
	}
	if d.warnings("op", nil) != nil {
		t.Fatal("no warnings must stay nil")
	}
}
