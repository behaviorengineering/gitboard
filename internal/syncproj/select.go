package syncproj

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// SelectInteractive runs a TUI multi-select (space toggle, enter confirm).
// Already-tracked repos start checked. Returns 1-based candidate indices.
func SelectInteractive(cands []Candidate) ([]int, error) {
	if len(cands) == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	byKey := map[string]int{}
	opts := make([]huh.Option[string], 0, len(cands))
	var selected []string
	for _, c := range cands {
		key := candidateKey(c)
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

func candidateKey(c Candidate) string {
	return string(c.Host) + "\x00" + strings.ToLower(strings.Trim(c.Path, "/"))
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
