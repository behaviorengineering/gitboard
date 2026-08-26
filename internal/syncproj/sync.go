package syncproj

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
)

// Lister discovers repositories from configured forges.
type Lister interface {
	ListGitHub(ctx context.Context, org string) ([]forge.RepoRef, error)
	ListGitLab(ctx context.Context, group string) ([]forge.RepoRef, error)
}

// ForgeLister adapts GitHub and GitLab clients.
type ForgeLister struct {
	GitHub *forge.GitHub
	GitLab *forge.GitLab
}

func (f ForgeLister) ListGitHub(ctx context.Context, org string) ([]forge.RepoRef, error) {
	return f.GitHub.ListOrgRepos(ctx, org)
}

func (f ForgeLister) ListGitLab(ctx context.Context, group string) ([]forge.RepoRef, error) {
	return f.GitLab.ListGroupRepos(ctx, group)
}

// Candidate is one selectable repository.
type Candidate struct {
	Host    config.Host
	Path    string
	Name    string
	Tracked bool
	Index   int // 1-based display index
}

// Discover lists unique repos from sync sources, marking already tracked ones.
func Discover(ctx context.Context, lister Lister, doc config.File, hostFilter string) ([]Candidate, error) {
	hostFilter = strings.ToLower(strings.TrimSpace(hostFilter))
	tracked := map[string]struct{}{}
	for _, p := range doc.Projects {
		tracked[trackKey(p.Host, p.Path)] = struct{}{}
	}

	seen := map[string]struct{}{}
	var refs []forge.RepoRef

	if hostFilter == "" || hostFilter == "github" {
		for _, org := range doc.Sync.GitHub.Orgs {
			list, err := lister.ListGitHub(ctx, org)
			if err != nil {
				return nil, fmt.Errorf("github org %s: %w", org, err)
			}
			refs = append(refs, list...)
		}
	}
	if hostFilter == "" || hostFilter == "gitlab" {
		for _, group := range doc.Sync.GitLab.Groups {
			list, err := lister.ListGitLab(ctx, group)
			if err != nil {
				return nil, fmt.Errorf("gitlab group %s: %w", group, err)
			}
			refs = append(refs, list...)
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Host != refs[j].Host {
			return refs[i].Host < refs[j].Host
		}
		return refs[i].Path < refs[j].Path
	})

	var out []Candidate
	for _, r := range refs {
		key := trackKey(r.Host, r.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		_, isTracked := tracked[key]
		out = append(out, Candidate{
			Host:    r.Host,
			Path:    r.Path,
			Name:    r.Name,
			Tracked: isTracked,
		})
	}
	for i := range out {
		out[i].Index = i + 1
	}
	return out, nil
}

// ProjectFromRef builds a config project from a forge repo.
func ProjectFromRef(host config.Host, path, name string) config.Project {
	path = strings.Trim(path, "/")
	id := slugID(name, path)
	label := name
	if label == "" {
		parts := strings.Split(path, "/")
		label = parts[len(parts)-1]
	}
	return config.Project{
		ID:    id,
		Label: label,
		Host:  host,
		Path:  path,
	}
}

// ApplySelection replaces projects with the selected candidates.
// Existing local_path values are kept when host+path still match.
func ApplySelection(cands []Candidate, selected []int, existing []config.Project) ([]config.Project, error) {
	byIndex := map[int]Candidate{}
	for _, c := range cands {
		byIndex[c.Index] = c
	}
	keepLocal := map[string]string{}
	for _, p := range existing {
		if strings.TrimSpace(p.LocalPath) == "" {
			continue
		}
		keepLocal[trackKey(p.Host, p.Path)] = p.LocalPath
	}
	var projects []config.Project
	usedIDs := map[string]struct{}{}
	for _, n := range selected {
		c, ok := byIndex[n]
		if !ok {
			return nil, fmt.Errorf("unknown selection index %d", n)
		}
		p := ProjectFromRef(c.Host, c.Path, c.Name)
		if lp, ok := keepLocal[trackKey(p.Host, p.Path)]; ok {
			p.LocalPath = lp
		}
		p.ID = uniqueID(p.ID, usedIDs)
		usedIDs[p.ID] = struct{}{}
		projects = append(projects, p)
	}
	return projects, nil
}

// AddProject appends or updates a project by host+path.
func AddProject(projects []config.Project, host config.Host, path string) ([]config.Project, error) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("path must be owner/repo")
	}
	name := parts[len(parts)-1]
	p := ProjectFromRef(host, path, name)
	key := trackKey(host, path)
	for i, existing := range projects {
		if trackKey(existing.Host, existing.Path) == key {
			p.ID = existing.ID
			p.LocalPath = existing.LocalPath
			projects[i] = p
			return projects, nil
		}
	}
	used := map[string]struct{}{}
	for _, existing := range projects {
		used[existing.ID] = struct{}{}
	}
	p.ID = uniqueID(p.ID, used)
	return append(projects, p), nil
}

func uniqueID(base string, used map[string]struct{}) string {
	if _, ok := used[base]; !ok {
		return base
	}
	for n := 2; ; n++ {
		id := fmt.Sprintf("%s-%d", base, n)
		if _, ok := used[id]; !ok {
			return id
		}
	}
}

// RemoveProject drops a project by id.
func RemoveProject(projects []config.Project, id string) ([]config.Project, bool) {
	id = strings.TrimSpace(id)
	var out []config.Project
	found := false
	for _, p := range projects {
		if p.ID == id {
			found = true
			continue
		}
		out = append(out, p)
	}
	return out, found
}

// ParseSelection parses "1,3,5", "all", or empty (keep current tracked indices).
func ParseSelection(line string, cands []Candidate) ([]int, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		var keep []int
		for _, c := range cands {
			if c.Tracked {
				keep = append(keep, c.Index)
			}
		}
		return keep, nil
	}
	if strings.EqualFold(line, "all") {
		out := make([]int, len(cands))
		for i, c := range cands {
			out[i] = c.Index
		}
		return out, nil
	}
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []int
	seen := map[int]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("bad selection %q", part)
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}

// FormatCandidates writes the numbered selection menu.
func FormatCandidates(w io.Writer, cands []Candidate) {
	for _, c := range cands {
		mark := " "
		if c.Tracked {
			mark = "x"
		}
		fmt.Fprintf(w, "[%d] [%s] %s %s\n", c.Index, mark, c.Host, c.Path)
	}
}

func trackKey(host config.Host, path string) string {
	return string(host) + ":" + strings.ToLower(strings.Trim(path, "/"))
}

func slugID(name, path string) string {
	base := name
	if base == "" {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		base = parts[len(parts)-1]
	}
	base = strings.ToLower(base)
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "repo"
	}
	return id
}
