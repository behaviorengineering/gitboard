package localgit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	headsPrefix  = "refs/heads/"
	originPrefix = "refs/remotes/origin/"
	originHEAD   = originPrefix + "HEAD"

	// refFormat prints, per ref, its full name, upstream, upstream track
	// ("ahead 1, behind 2", "gone" or empty) and symref target, NUL-separated.
	refFormat = "%(refname)%00%(upstream)%00%(upstream:track,nobracket)%00%(symref)"
	// aheadBehindAtom adds "<ahead> <behind>" against the origin default
	// branch; git 2.41 added the atom.
	aheadBehindAtom = "%00%(ahead-behind:" + originHEAD + ")"

	// originCompareParallel caps the rev-list fallbacks for branches whose
	// upstream is not origin/<same name>.
	originCompareParallel = 8
)

// RepoSync is origin sync state for one repository. Refs live in the common
// git dir, so every worktree of a repository shares one RepoSync.
type RepoSync struct {
	// DefaultBranch is the branch origin/HEAD names (else main or master when
	// origin has them); empty when unknown.
	DefaultBranch string
	// Branches lists local heads that have an origin twin, sorted by name.
	Branches []BranchSync
}

// Apply copies r onto st: OriginSync, DefaultBranch when known, and the
// default branch's ahead and behind counts.
func (r RepoSync) Apply(st *Status) {
	if st == nil {
		return
	}
	if r.DefaultBranch != "" {
		st.DefaultBranch = r.DefaultBranch
	}
	if len(r.Branches) == 0 {
		st.OriginSync = nil
		return
	}
	st.OriginSync = append([]BranchSync(nil), r.Branches...)
	for _, b := range r.Branches {
		if b.Name == st.DefaultBranch {
			st.DefaultAhead = b.Ahead
			st.DefaultBehind = b.Behind
		}
	}
}

// EnrichOriginSync runs OriginSync for st.Path and applies it. A failure is
// added to st.Error and leaves OriginSync empty. Callers with several
// worktrees of one repository should call OriginSync once and Apply it to
// each Status instead.
func (in *Inspector) EnrichOriginSync(ctx context.Context, st *Status) {
	if in == nil || st == nil || !st.Mapped || st.Path == "" || st.Error != "" {
		return
	}
	rs, err := in.OriginSync(ctx, st.Path)
	rs.Apply(st)
	if err != nil {
		st.Error = err.Error()
		InvalidateOriginSync(st)
	}
}

// OriginSync compares each local branch to origin/<same name> for the
// repository at dir. One for-each-ref lists local and origin refs with their
// upstream track counts; a branch whose upstream is origin/<same name> uses
// those counts, and only the others run rev-list. On git 2.41 or newer the
// same call also gives the default branch's position, so it never needs a
// rev-list either.
func (in *Inspector) OriginSync(ctx context.Context, dir string) (RepoSync, error) {
	refs, err := in.listOriginRefs(ctx, dir)
	if err != nil {
		return RepoSync{}, fmt.Errorf("list branches: %w", err)
	}
	out := RepoSync{DefaultBranch: refs.defaultBranch()}
	syncs, err := in.branchSyncs(ctx, dir, refs)
	if err != nil {
		return out, fmt.Errorf("could not compare local branches to origin: %w", err)
	}
	out.Branches = syncs
	return out, nil
}

// refRow is one line of refFormat output.
type refRow struct {
	name      string // full ref name.
	upstream  string // full upstream ref name; empty when none.
	track     string // upstream:track,nobracket.
	symref    string // symref target; set for origin/HEAD.
	vsDefault string // "<ahead> <behind>" against origin/HEAD; empty when not asked.
}

// originRefs is the parsed for-each-ref listing for one repository.
type originRefs struct {
	heads      []refRow            // local branches (short names in name).
	remote     map[string]struct{} // origin branch short names, HEAD excluded.
	originHead string              // branch origin/HEAD points at; empty when unset.
}

func (in *Inspector) listOriginRefs(ctx context.Context, dir string) (originRefs, error) {
	if in.supportsAheadBehind(ctx) {
		out, err := in.git(ctx, dir, "for-each-ref", "--format="+refFormat+aheadBehindAtom, headsPrefix, originPrefix)
		if err == nil {
			return parseOriginRefs(out)
		}
		// origin/HEAD is missing (the atom then fails the whole call); list
		// again without it.
	}
	out, err := in.git(ctx, dir, "for-each-ref", "--format="+refFormat, headsPrefix, originPrefix)
	if err != nil {
		return originRefs{}, err
	}
	return parseOriginRefs(out)
}

func parseOriginRefs(out []byte) (originRefs, error) {
	refs := originRefs{remote: map[string]struct{}{}}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		row := parseRefRow(line)
		switch {
		case row.name == originHEAD:
			refs.originHead = strings.TrimPrefix(row.symref, originPrefix)
		case strings.HasPrefix(row.name, originPrefix):
			refs.remote[strings.TrimPrefix(row.name, originPrefix)] = struct{}{}
		case strings.HasPrefix(row.name, headsPrefix):
			row.name = strings.TrimPrefix(row.name, headsPrefix)
			refs.heads = append(refs.heads, row)
		}
	}
	if err := sc.Err(); err != nil {
		return originRefs{}, err
	}
	return refs, nil
}

func parseRefRow(line string) refRow {
	f := strings.Split(line, "\x00")
	for len(f) < 5 {
		f = append(f, "")
	}
	return refRow{name: f[0], upstream: f[1], track: f[2], symref: f[3], vsDefault: f[4]}
}

// defaultBranch is origin/HEAD's target, else main or master when origin has it.
func (r originRefs) defaultBranch() string {
	if r.originHead != "" {
		return r.originHead
	}
	for _, name := range []string{defaultBranchFallback, "master"} {
		if _, ok := r.remote[name]; ok {
			return name
		}
	}
	return ""
}

// branchSyncs returns one BranchSync per local head with an origin twin,
// sorted by name. Heads tracking their twin use the listed counts.
func (in *Inspector) branchSyncs(ctx context.Context, dir string, refs originRefs) ([]BranchSync, error) {
	var out []BranchSync
	var compare []string
	for _, h := range refs.heads {
		if _, ok := refs.remote[h.name]; !ok {
			continue
		}
		if s, ok := h.listedSync(refs.originHead); ok {
			out = append(out, s)
			continue
		}
		compare = append(compare, h.name)
	}
	compared, err := in.compareTwins(ctx, dir, compare)
	if err != nil {
		return nil, err
	}
	out = append(out, compared...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// listedSync returns h's counts against origin/<name> from the listing:
// upstream track counts when h tracks origin/<name>, or the ahead-behind
// atom when h is the default branch. ok is false when neither applies.
func (h refRow) listedSync(originHead string) (BranchSync, bool) {
	if h.upstream == originPrefix+h.name {
		if ahead, behind, err := parseTrack(h.track); err == nil {
			return BranchSync{Name: h.name, Ahead: ahead, Behind: behind}, true
		}
	}
	if h.name == originHead && h.vsDefault != "" {
		if ahead, behind, err := parseAheadBehind(h.vsDefault); err == nil {
			return BranchSync{Name: h.name, Ahead: ahead, Behind: behind}, true
		}
	}
	return BranchSync{}, false
}

// errUpstreamGone means the upstream ref was deleted (track says "gone").
var errUpstreamGone = errors.New("upstream gone")

// parseTrack reads %(upstream:track,nobracket): "" (in sync), "ahead N",
// "behind N", "ahead N, behind M" or "gone".
func parseTrack(track string) (ahead, behind int, err error) {
	track = strings.TrimSpace(track)
	if track == "" {
		return 0, 0, nil
	}
	if track == "gone" {
		return 0, 0, errUpstreamGone
	}
	for _, part := range strings.Split(track, ",") {
		key, val, ok := strings.Cut(strings.TrimSpace(part), " ")
		if !ok {
			return 0, 0, fmt.Errorf("track %q: no count", track)
		}
		n, convErr := strconv.Atoi(strings.TrimSpace(val))
		if convErr != nil {
			return 0, 0, fmt.Errorf("track %q: %w", track, convErr)
		}
		switch key {
		case "ahead":
			ahead = n
		case "behind":
			behind = n
		default:
			return 0, 0, fmt.Errorf("track %q: unknown key %q", track, key)
		}
	}
	return ahead, behind, nil
}

// parseAheadBehind reads %(ahead-behind:...): "<ahead> <behind>".
func parseAheadBehind(value string) (ahead, behind int, err error) {
	// parseLeftRightCount reads "<left> <right>"; here left is ahead.
	ahead, behind, err = parseLeftRightCount(value)
	if err != nil {
		return 0, 0, fmt.Errorf("ahead-behind: %w", err)
	}
	return ahead, behind, nil
}

// compareTwins runs rev-list for each named branch against origin/<name>,
// at most originCompareParallel at once. Any failure fails the whole set.
func (in *Inspector) compareTwins(ctx context.Context, dir string, names []string) ([]BranchSync, error) {
	if len(names) == 0 {
		return nil, nil
	}
	syncs := make([]BranchSync, len(names))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(originCompareParallel)
	for i, name := range names {
		g.Go(func() error {
			ahead, behind, err := in.leftRight(gctx, dir, originPrefix+name, headsPrefix+name)
			if err != nil {
				return err
			}
			syncs[i] = BranchSync{Name: name, Ahead: ahead, Behind: behind}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return syncs, nil
}

// minAheadBehind is the first git release with %(ahead-behind:...).
var minAheadBehind = gitVersion{major: 2, minor: 41}

type gitVersion struct{ major, minor int }

func (v gitVersion) atLeast(o gitVersion) bool {
	return v.major > o.major || (v.major == o.major && v.minor >= o.minor)
}

// parseGitVersion reads "git version 2.46.2" or "git version 2.39.3 (Apple Git-146)".
func parseGitVersion(out string) (gitVersion, bool) {
	fields := strings.Fields(out)
	if len(fields) < 3 || fields[0] != "git" || fields[1] != "version" {
		return gitVersion{}, false
	}
	parts := strings.SplitN(fields[2], ".", 3)
	if len(parts) < 2 {
		return gitVersion{}, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return gitVersion{}, false
	}
	return gitVersion{major: major, minor: minor}, true
}

// versionProbe runs `git version` once for an Inspector. NewInspector
// creates one, so the dashboard's shared Inspector probes once per process.
type versionProbe struct {
	mu      sync.Mutex
	done    bool
	version gitVersion
	ok      bool
}

// supportsAheadBehind reports whether git has %(ahead-behind:...). An
// Inspector built without NewInspector, or a failed probe, answers false.
// A probe cut short by ctx is retried on the next call.
func (in *Inspector) supportsAheadBehind(ctx context.Context) bool {
	p := in.version
	if p == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.done {
		out, err := in.git(ctx, ".", "version")
		if err != nil && ctx.Err() != nil {
			return false
		}
		p.done = true
		if err == nil {
			p.version, p.ok = parseGitVersion(string(out))
		}
	}
	return p.ok && p.version.atLeast(minAheadBehind)
}
