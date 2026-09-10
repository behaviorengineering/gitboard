package localgit

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const syncCommitLimit = 20

// SyncCommit is one commit that exists on only one side of a local/origin comparison.
type SyncCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// SyncInspection describes the current relationship between a local branch and origin.
type SyncInspection struct {
	Path           string       `json:"path"`
	Branch         string       `json:"branch"`
	CurrentBranch  string       `json:"current_branch,omitempty"`
	Upstream       string       `json:"upstream,omitempty"`
	LocalSHA       string       `json:"local_sha,omitempty"`
	RemoteSHA      string       `json:"remote_sha,omitempty"`
	MergeBase      string       `json:"merge_base,omitempty"`
	Relation       string       `json:"relation"`
	Dirty          bool         `json:"dirty"`
	DirtyFiles     []string     `json:"dirty_files,omitempty"`
	AheadCount     int          `json:"ahead_count"`
	BehindCount    int          `json:"behind_count"`
	Ahead          []SyncCommit `json:"ahead,omitempty"`
	Behind         []SyncCommit `json:"behind,omitempty"`
	CanFastForward bool         `json:"can_fast_forward"`
}

// InspectSync refreshes origin and compares a local branch with origin/<branch>.
func (in *Inspector) InspectSync(ctx context.Context, repoPath, branch string) (SyncInspection, error) {
	var out SyncInspection
	if in == nil {
		return out, fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return out, err
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return out, fmt.Errorf("expand path: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(abs); resolveErr == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if !in.isGitDir(ctx, abs) {
		return out, fmt.Errorf("not a git repository: %s", abs)
	}

	out.Path = abs
	out.Branch = branch
	out.Upstream = "origin/" + branch
	if _, err := in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return out, fmt.Errorf("fetch --prune origin: %w", err)
	}

	localRef := "refs/heads/" + branch
	remoteRef := "refs/remotes/origin/" + branch
	out.LocalSHA, err = in.revParse(ctx, abs, localRef)
	if err != nil {
		return out, fmt.Errorf("local branch %q: %w", branch, err)
	}
	out.RemoteSHA, err = in.revParse(ctx, abs, remoteRef)
	if err != nil {
		return out, fmt.Errorf("origin branch %q: %w", branch, err)
	}

	current, currentErr := in.git(ctx, abs, "rev-parse", "--abbrev-ref", "HEAD")
	if currentErr != nil {
		return out, fmt.Errorf("current branch: %w", currentErr)
	}
	out.CurrentBranch = strings.TrimSpace(string(current))
	status, statusErr := in.git(ctx, abs, "status", "--porcelain")
	if statusErr != nil {
		return out, fmt.Errorf("status: %w", statusErr)
	}
	if out.CurrentBranch == branch {
		out.DirtyFiles = porcelainFiles(string(status))
		out.Dirty = len(out.DirtyFiles) > 0
	}

	counts, countsErr := in.git(ctx, abs, "rev-list", "--left-right", "--count", remoteRef+"..."+localRef)
	if countsErr != nil {
		return out, fmt.Errorf("compare branch %q with origin: %w", branch, countsErr)
	}
	out.BehindCount, out.AheadCount, err = parseLeftRightCount(string(counts))
	if err != nil {
		return out, fmt.Errorf("compare branch %q with origin: %w", branch, err)
	}

	if base, baseErr := in.git(ctx, abs, "merge-base", localRef, remoteRef); baseErr == nil {
		out.MergeBase = strings.TrimSpace(string(base))
	} else {
		out.Relation = "unrelated"
		return out, nil
	}

	out.Behind, err = in.syncCommits(ctx, abs, localRef+".."+remoteRef)
	if err != nil {
		return out, fmt.Errorf("list commits behind %q: %w", branch, err)
	}
	out.Ahead, err = in.syncCommits(ctx, abs, remoteRef+".."+localRef)
	if err != nil {
		return out, fmt.Errorf("list commits ahead of %q: %w", branch, err)
	}

	switch {
	case out.AheadCount == 0 && out.BehindCount == 0:
		out.Relation = "up_to_date"
	case out.AheadCount == 0:
		out.Relation = "behind_only"
		out.CanFastForward = !out.Dirty
	case out.BehindCount == 0:
		out.Relation = "ahead_only"
	default:
		out.Relation = "diverged"
	}
	return out, nil
}

func (in *Inspector) revParse(ctx context.Context, dir, ref string) (string, error) {
	out, err := in.git(ctx, dir, "rev-parse", "--verify", ref)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", fmt.Errorf("empty ref")
	}
	return value, nil
}

func (in *Inspector) syncCommits(ctx context.Context, dir, rangeSpec string) ([]SyncCommit, error) {
	out, err := in.git(ctx, dir, "log", "--max-count", fmt.Sprint(syncCommitLimit), "--format=%h%x09%s", rangeSpec)
	if err != nil {
		return nil, err
	}
	var commits []SyncCommit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sha, subject, ok := strings.Cut(line, "\t")
		if !ok {
			commits = append(commits, SyncCommit{SHA: line})
			continue
		}
		commits = append(commits, SyncCommit{
			SHA:     strings.TrimSpace(sha),
			Subject: strings.TrimSpace(subject),
		})
	}
	return commits, nil
}

func porcelainFiles(status string) []string {
	var files []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func parseLeftRightCount(value string) (behind, ahead int, err error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid count %q", strings.TrimSpace(value))
	}
	if _, err := fmt.Sscanf(parts[0], "%d", &behind); err != nil {
		return 0, 0, fmt.Errorf("invalid behind count %q: %w", parts[0], err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &ahead); err != nil {
		return 0, 0, fmt.Errorf("invalid ahead count %q: %w", parts[1], err)
	}
	return behind, ahead, nil
}
