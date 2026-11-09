package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	ollycli "github.com/behaviorengineering/olly/pkg/cli"
	"github.com/spf13/cobra"

	"github.com/behaviorengineering/gitboard/internal/cli/syncprompt"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/projectsync"
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
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := &syncRun{
				f:      f,
				out:    &printer{w: cmd.OutOrStdout()},
				errOut: cmd.ErrOrStderr(),
				log:    observability.NewLogger().With("cmd", "sync"),
			}
			return s.run(cmd.Context(), interactive)
		},
	}
	cmd.Flags().StringVar(&f.path, "config", config.DefaultPath(), configFlagHelp)
	cmd.Flags().StringVar(&f.host, "host", "", "limit discovery to github, gitlab, azuredevops, or bitbucket")
	cmd.Flags().StringVar(&f.add, "add", "", "non-interactive: add owner/repo (requires --host)")
	cmd.Flags().StringVar(&f.remove, "remove", "", "non-interactive: remove project by id")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "discover and print what would change; writes nothing, not even a missing config file")
	return cmd
}

// syncRun is one gitboard sync invocation. Every config rule goes through
// the same dashboard.Commands the Manage API uses; sync only prompts, prints
// and saves.
type syncRun struct {
	f      syncFlags
	out    *printer
	errOut io.Writer
	log    observability.Logger
	cmds   *dashboard.Commands
}

func (s *syncRun) run(ctx context.Context, interactive func() bool) (err error) {
	removing := strings.TrimSpace(s.f.remove) != ""
	adding := strings.TrimSpace(s.f.add) != ""
	if !removing && !adding && !s.f.dryRun && !interactive() {
		return usage(errNeedsTerminal)
	}
	doc, err := s.loadOrCreate()
	if err != nil {
		return err
	}
	// Sync makes forge calls, so it installs tracing like serve and runs
	// under one command span. It writes no failure dumps: those are for LLM
	// work, and a dry run must not create files.
	tr := newTracing(doc, s.log, false)
	stop, err := tr.start(ctx, s.log)
	if err != nil {
		return err
	}
	defer stop()
	ctx, span := ollycli.WithCommandSpan(ctx, tr.serviceName(), "gitboard sync")
	defer func() { ollycli.EndWithErr(span, err) }()

	s.cmds = dashboard.NewCommands(dashboardConfig(newRunner(), time.Now).CreateService())
	switch {
	case removing:
		return s.remove(doc)
	case adding:
		return s.addProject(doc)
	default:
		return s.discover(ctx, doc)
	}
}

// loadOrCreate loads the config, creating it from the template first unless
// this is a dry run. A dry run against a missing file works on an empty
// document and says what it would have created.
func (s *syncRun) loadOrCreate() (boardconfig.File, error) {
	if config.Exists(s.f.path) {
		return loadConfig(s.f.path)
	}
	if s.f.dryRun {
		s.out.printf("dry-run: would create %s\n", s.f.path)
		return boardconfig.File{}, s.out.err
	}
	if _, err := config.Init(s.f.path); err != nil {
		return boardconfig.File{}, err
	}
	s.out.printf("created %s\n", s.f.path)
	return loadConfig(s.f.path)
}

func (s *syncRun) remove(doc boardconfig.File) error {
	updated, err := s.cmds.RemoveTrackedProject(doc, s.f.remove)
	if err != nil {
		return err
	}
	if s.f.dryRun {
		s.out.printf("dry-run: would remove %s (%d projects left)\n", s.f.remove, len(updated.Projects))
		return s.out.err
	}
	if err := config.Save(s.f.path, updated); err != nil {
		return err
	}
	s.out.printf("removed %s (%d projects)\n", s.f.remove, len(updated.Projects))
	return s.out.err
}

func (s *syncRun) addProject(doc boardconfig.File) error {
	host := boardconfig.Host(strings.ToLower(strings.TrimSpace(s.f.host)))
	updated, err := s.cmds.AddTrackedProject(doc, host, s.f.add)
	if err != nil {
		return err
	}
	if s.f.dryRun {
		s.out.printf("dry-run: would add %s %s (%d projects)\n", host, s.f.add, len(updated.Projects))
		return s.out.err
	}
	if err := config.Save(s.f.path, updated); err != nil {
		return err
	}
	s.out.printf("tracked %s %s (%d projects)\n", host, s.f.add, len(updated.Projects))
	return s.out.err
}

func (s *syncRun) discover(ctx context.Context, doc boardconfig.File) error {
	doc, err := s.ensureSources(doc)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, syncBudget)
	defer cancel()
	res, err := s.cmds.DiscoverCandidates(ctx, doc, s.f.host)
	if err != nil {
		return err
	}
	ew := &printer{w: s.errOut}
	for _, w := range res.Warnings {
		ew.printf("warning: %s\n", w)
	}
	if ew.err != nil {
		return ew.err
	}
	cands := res.Candidates
	if len(cands) == 0 {
		if len(res.Warnings) > 0 {
			return gberr.New(gberr.CodeUnavailable, "cli.sync", fmt.Sprintf("no repositories found; %d sync source(s) failed", len(res.Warnings)))
		}
		return gberr.New(gberr.CodeNotFound, "cli.sync", "no repositories found for configured sync sources")
	}
	s.out.printf("Candidates: %d repositories\n", len(cands))
	if s.f.dryRun {
		syncprompt.FormatCandidates(s.out.w, cands)
		s.out.printf("dry-run: not writing config\n")
		return s.out.err
	}
	return s.selectAndSave(doc, cands)
}

// ensureSources prompts for orgs and groups when the config has none.
func (s *syncRun) ensureSources(doc boardconfig.File) (boardconfig.File, error) {
	if doc.Sync.HasSyncSources() {
		return doc, nil
	}
	if s.f.dryRun {
		return doc, gberr.New(gberr.CodeInvalid, "cli.sync",
			fmt.Sprintf("no sync sources: set sync.github.orgs / sync.gitlab.groups in %s (or run sync without --dry-run)", s.f.path))
	}
	s.out.printf("No sync sources configured.\n")
	orgs, groups, err := syncprompt.PromptSyncSources()
	if err != nil {
		return doc, err
	}
	doc, err = s.cmds.SetSyncSources(doc, dashboard.SyncSourceLists{
		GitHubOrgs:          orgs,
		GitLabGroups:        groups,
		AzureDevOpsOrgs:     doc.Sync.AzureDevOps.Orgs,
		BitbucketWorkspaces: doc.Sync.Bitbucket.Workspaces,
	})
	if err != nil {
		return doc, err
	}
	if err := config.Save(s.f.path, doc); err != nil {
		return doc, err
	}
	s.out.printf("saved sync sources to %s\n", s.f.path)
	return doc, nil
}

func (s *syncRun) selectAndSave(doc boardconfig.File, cands []projectsync.Candidate) error {
	selected, err := syncprompt.SelectInteractive(cands)
	if err != nil {
		return err
	}
	byIndex := make(map[int]projectsync.Candidate, len(cands))
	for _, c := range cands {
		byIndex[c.Index] = c
	}
	refs := make([]projectsync.RepoRef, 0, len(selected))
	for _, idx := range selected {
		if c, ok := byIndex[idx]; ok {
			refs = append(refs, projectsync.RepoRef{Host: c.Host, Path: c.Path})
		}
	}
	doc, err = s.cmds.ApplySelection(doc, cands, refs)
	if err != nil {
		return err
	}
	if err := config.Save(s.f.path, doc); err != nil {
		return err
	}
	s.out.printf("wrote %d projects to %s\n", len(doc.Projects), s.f.path)
	return s.out.err
}
