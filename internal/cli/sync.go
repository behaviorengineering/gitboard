package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// syncBudget bounds forge discovery for one sync run.
const syncBudget = 3 * time.Minute

// errNeedsTerminal is returned when interactive sync runs without a terminal.
var errNeedsTerminal = errors.New("sync needs a terminal for interactive selection; " +
	"use --host and --add owner/repo, --remove <id>, or --dry-run")

type syncFlags struct {
	path   string
	host   string
	add    string
	remove string
	dryRun bool
}

func newSyncCmd(interactive func() bool) *cobra.Command {
	var f syncFlags
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Discover repos and select tracked projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSync(cmd.Context(), &printer{w: cmd.OutOrStdout()}, cmd.ErrOrStderr(), f, interactive)
		},
	}
	cmd.Flags().StringVar(&f.path, "config", config.DefaultPath(), "config file path (GITBOARD_CONFIG, then ~/.config/gitboard/config.yaml)")
	cmd.Flags().StringVar(&f.host, "host", "", "limit discovery to github, gitlab, azuredevops, or bitbucket")
	cmd.Flags().StringVar(&f.add, "add", "", "non-interactive: add owner/repo (requires --host)")
	cmd.Flags().StringVar(&f.remove, "remove", "", "non-interactive: remove project by id")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "discover and print candidates without writing")
	return cmd
}

func runSync(ctx context.Context, p *printer, errOut io.Writer, f syncFlags, interactive func() bool) error {
	removing := strings.TrimSpace(f.remove) != ""
	adding := strings.TrimSpace(f.add) != ""
	if !removing && !adding && !f.dryRun && !interactive() {
		return errNeedsTerminal
	}
	if !fileExists(f.path) {
		if _, err := config.Init(f.path); err != nil {
			return err
		}
		p.printf("created %s\n", f.path)
	}
	doc, err := config.Load(f.path)
	if err != nil {
		return err
	}
	switch {
	case removing:
		return syncRemove(p, f, doc)
	case adding:
		return syncAdd(p, f, doc)
	default:
		return syncDiscover(ctx, p, errOut, f, doc)
	}
}

func syncRemove(p *printer, f syncFlags, doc config.File) error {
	updated, err := syncCommands().RemoveTrackedProject(doc, f.remove)
	if err != nil {
		return err
	}
	if f.dryRun {
		p.printf("dry-run: would remove %s (%d projects left)\n", f.remove, len(updated.Projects))
		return p.err
	}
	if err := config.Save(f.path, updated); err != nil {
		return err
	}
	p.printf("removed %s (%d projects)\n", f.remove, len(updated.Projects))
	return p.err
}

func syncAdd(p *printer, f syncFlags, doc config.File) error {
	host := config.Host(strings.ToLower(strings.TrimSpace(f.host)))
	updated, err := syncCommands().AddTrackedProject(doc, host, f.add)
	if err != nil {
		return err
	}
	if f.dryRun {
		p.printf("dry-run: would add %s %s (%d projects)\n", host, f.add, len(updated.Projects))
		return p.err
	}
	if err := config.Save(f.path, updated); err != nil {
		return err
	}
	p.printf("tracked %s %s (%d projects)\n", host, f.add, len(updated.Projects))
	return p.err
}

func syncDiscover(ctx context.Context, p *printer, errOut io.Writer, f syncFlags, doc config.File) error {
	doc, err := ensureSyncSources(p, f, doc)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, syncBudget)
	defer cancel()
	res, err := syncCommands().DiscoverCandidates(ctx, doc, f.host)
	if err != nil {
		return err
	}
	ew := &printer{w: errOut}
	for _, w := range res.Warnings {
		ew.printf("warning: %s\n", w)
	}
	if ew.err != nil {
		return ew.err
	}
	cands := res.Candidates
	if len(cands) == 0 {
		if len(res.Warnings) > 0 {
			return fmt.Errorf("no repositories found; %d sync source(s) failed", len(res.Warnings))
		}
		return errors.New("no repositories found for configured sync sources")
	}
	p.printf("Candidates: %d repositories\n", len(cands))
	if f.dryRun {
		syncproj.FormatCandidates(p.w, cands)
		p.printf("dry-run: not writing config\n")
		return p.err
	}
	return selectAndSave(p, f.path, doc, cands)
}

// ensureSyncSources prompts for orgs and groups when the config has none.
func ensureSyncSources(p *printer, f syncFlags, doc config.File) (config.File, error) {
	if doc.Sync.HasSyncSources() {
		return doc, nil
	}
	if f.dryRun {
		return doc, fmt.Errorf("no sync sources: set sync.github.orgs / sync.gitlab.groups in %s (or run sync without --dry-run)", f.path)
	}
	p.printf("No sync sources configured.\n")
	orgs, groups, err := syncproj.PromptSyncSources()
	if err != nil {
		return doc, err
	}
	doc.Sync.GitHub.Orgs = orgs
	doc.Sync.GitLab.Groups = groups
	if !doc.Sync.HasSyncSources() {
		return doc, errors.New("need at least one github org or gitlab group")
	}
	if err := config.Save(f.path, doc); err != nil {
		return doc, err
	}
	p.printf("saved sync sources to %s\n", f.path)
	return doc, nil
}

func selectAndSave(p *printer, path string, doc config.File, cands []syncproj.Candidate) error {
	selected, err := syncproj.SelectInteractive(cands)
	if err != nil {
		return err
	}
	byIndex := make(map[int]syncproj.Candidate, len(cands))
	for _, c := range cands {
		byIndex[c.Index] = c
	}
	refs := make([]syncproj.RepoRef, 0, len(selected))
	for _, idx := range selected {
		if c, ok := byIndex[idx]; ok {
			refs = append(refs, syncproj.RepoRef{Host: c.Host, Path: c.Path})
		}
	}
	projects, err := syncproj.ApplySelectionByRefs(cands, refs, doc.Projects)
	if err != nil {
		return err
	}
	doc.Projects = projects
	doc.Views = config.PruneViewMembership(doc.Views, doc.Projects)
	if err := config.Save(path, doc); err != nil {
		return err
	}
	p.printf("wrote %d projects to %s\n", len(doc.Projects), path)
	return p.err
}

// syncCommands builds the same Commands surface the Manage API uses for sync mutations.
func syncCommands() *dashboard.Commands {
	run := cliexec.New()
	run.Timeout = 120 * time.Second
	dash := dashboard.New(
		remotegit.NewGitHub(run),
		remotegit.NewGitLab(run),
		remotegit.NewAzureDevOps(run),
		remotegit.NewBitbucket(run),
		nil,
	)
	return dashboard.NewCommands(dash)
}
