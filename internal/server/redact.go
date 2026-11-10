package server

import (
	"fmt"

	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/board"
)

// Display text the UI shows in place of raw error detail on a remote server.
const (
	redactedProjectError = "forge data unavailable; details are in the server log"
	redactedLocalError   = "local checkout check failed; details are in the server log"
)

// detailRedactor trims error and detail text before it leaves a server that
// may have remote readers. That text can hold local paths and git or forge
// CLI stderr. On loopback (on false) it passes everything through unchanged.
// Paths shown as board content (checkout and worktree paths) are kept.
type detailRedactor struct {
	on  bool
	log observability.Logger
}

// project trims the forge error and every local error on one row.
func (d detailRedactor) project(row *board.ProjectSummary) {
	if !d.on || row == nil {
		return
	}
	log := d.log.With("project", row.ID)
	row.Error = d.swap(log, "project.error", row.Error, redactedProjectError)
	if row.Local == nil {
		return
	}
	row.Local.Error = d.swap(log, "local.error", row.Local.Error, redactedLocalError)
	for i := range row.Local.Appearances {
		app := &row.Local.Appearances[i]
		app.Error = d.swap(log.With("path", app.Path), "appearance.error", app.Error, redactedLocalError)
	}
}

// projects trims every row in place.
func (d detailRedactor) projects(rows []board.ProjectSummary) {
	for i := range rows {
		d.project(&rows[i])
	}
}

// tooling drops the per-forge detail text; the installed and authed flags stay.
func (d detailRedactor) tooling(t *board.Tooling) {
	if !d.on || t == nil {
		return
	}
	t.GitHub.Detail = d.swap(d.log.With("forge", "github"), "tooling.detail", t.GitHub.Detail, "")
	t.GitLab.Detail = d.swap(d.log.With("forge", "gitlab"), "tooling.detail", t.GitLab.Detail, "")
	t.AzureDevOps.Detail = d.swap(d.log.With("forge", "azuredevops"), "tooling.detail", t.AzureDevOps.Detail, "")
	t.Bitbucket.Detail = d.swap(d.log.With("forge", "bitbucket"), "tooling.detail", t.Bitbucket.Detail, "")
}

// warnings replaces a list of warning lines with one count line.
func (d detailRedactor) warnings(op string, lines []string) []string {
	if !d.on || len(lines) == 0 {
		return lines
	}
	for _, line := range lines {
		d.log.With("op", op).Infof("redacted warning: %s", line)
	}
	return []string{fmt.Sprintf("%d warnings; details are in the server log", len(lines))}
}

// swap logs text and returns replacement; empty text stays empty.
func (d detailRedactor) swap(log observability.Logger, field, text, replacement string) string {
	if text == "" {
		return ""
	}
	log.With("field", field).Infof("redacted display text: %s", text)
	return replacement
}
