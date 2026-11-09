// Package syncprompt is the interactive terminal surface of gitboard sync:
// the numbered candidate list, the multi-select and the sync source prompt.
package syncprompt

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
)

// SelectInteractive runs a TUI multi-select (space toggle, enter confirm).
// Already-tracked repos start checked. Returns 1-based candidate indices.
func SelectInteractive(cands []projectsync.Candidate) ([]int, error) {
	if len(cands) == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	byKey := map[string]int{}
	opts := make([]huh.Option[string], 0, len(cands))
	var selected []string
	for _, c := range cands {
		key := localgit.TrackKey(c.Host, c.Path)
		byKey[key] = c.Index
		label := fmt.Sprintf("%s  %s", c.Host, c.Path)
		opts = append(opts, huh.NewOption(label, key))
		if c.Tracked {
			selected = append(selected, key)
		}
	}
	height := len(cands) + 2
	if height > 18 {
		height = 18
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Track repositories").
				Description("↑↓ move · space toggle · / filter · enter confirm").
				Options(opts...).
				Filterable(true).
				Height(height).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	out := make([]int, 0, len(selected))
	seen := map[int]struct{}{}
	for _, key := range selected {
		idx, ok := byKey[key]
		if !ok {
			continue
		}
		if _, dup := seen[idx]; dup {
			continue
		}
		seen[idx] = struct{}{}
		out = append(out, idx)
	}
	return out, nil
}

// PromptSyncSources asks for GitHub orgs and GitLab groups via TUI inputs.
func PromptSyncSources() (orgs, groups []string, err error) {
	var orgLine, groupLine string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub organizations").
				Description("Comma-separated (empty to skip)").
				Placeholder("behaviorengineering, xynova").
				Value(&orgLine),
			huh.NewInput().
				Title("GitLab groups").
				Description("Comma-separated (empty to skip)").
				Placeholder("behaviorengineering").
				Value(&groupLine),
		),
	)
	if err := form.Run(); err != nil {
		return nil, nil, err
	}
	return splitCSV(orgLine), splitCSV(groupLine), nil
}

func splitCSV(line string) []string {
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// FormatCandidates writes the numbered selection menu.
func FormatCandidates(w io.Writer, cands []projectsync.Candidate) {
	for _, c := range cands {
		mark := " "
		if c.Tracked {
			mark = "x"
		}
		if _, err := fmt.Fprintf(w, "[%d] [%s] %s %s\n", c.Index, mark, c.Host, c.Path); err != nil {
			return
		}
	}
}
