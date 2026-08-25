// Local code-change board for GitLab and GitHub (127.0.0.1 only).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/dashboard"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/localgit"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/internal/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/internal/syncproj"
	"github.com/behaviorengineering/gitboard/internal/triage"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		if err := runServe(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "init":
		err = runInit(args)
	case "sync":
		err = runSync(args)
	case "serve":
		err = runServe(args)
	case "-h", "--help", "help":
		printUsage(os.Stdout)
	default:
		// Backward compatible: treat unknown first arg as serve flags (e.g. -addr).
		if strings.HasPrefix(cmd, "-") {
			err = runServe(os.Args[1:])
		} else {
			fmt.Fprintf(os.Stderr, "gitboard: unknown command %q\n\n", cmd)
			printUsage(os.Stderr)
			os.Exit(2)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `gitboard — GitLab + GitHub code-change board (gh + glab)

Usage:
  gitboard init
  gitboard sync [flags]
  gitboard serve [flags]

Config: %s
`, config.DefaultPath())
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	path := fs.String("config", config.DefaultPath(), "config file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	created, err := config.Init(*path)
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("created %s\n", *path)
	} else {
		fmt.Printf("already exists: %s (not overwritten)\n", *path)
	}
	fmt.Println("next: edit llm settings if needed, then: gitboard sync")
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultPath(), "config file path")
	projectsPath := fs.String("projects", "", "legacy alias for -config")
	addr := fs.String("addr", "127.0.0.1:1325", "listen address (localhost only)")
	allowNonLocalhost := fs.Bool("allow-non-localhost", false, "allow bind on non-loopback")
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := *configPath
	if strings.TrimSpace(*projectsPath) != "" {
		path = *projectsPath
	}
	if !*allowNonLocalhost && !strings.HasPrefix(*addr, "127.0.0.1:") && !strings.HasPrefix(*addr, "localhost:") {
		return fmt.Errorf("refuse to bind outside localhost (use -allow-non-localhost to override)")
	}
	doc, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("%w\nrun: gitboard init && gitboard sync", err)
	}
	oi := doc.EffectiveOpenInference()
	var dumpDir string
	if oi.Enabled != nil && *oi.Enabled && oi.FailureDump.Enabled != nil && *oi.FailureDump.Enabled {
		dumpDir = oi.FailureDump.Dir
	}
	tp, err := observability.Init(observability.InitConfig{
		ServiceName:            oi.ServiceName,
		OTLPEndpoint:           oi.Endpoint,
		FailureDumpDir:         dumpDir,
		FailureDumpMaxAgeHours: oi.FailureDump.MaxAgeHours,
		FailureDumpMaxFiles:    oi.FailureDump.MaxFiles,
	})
	if err != nil {
		return fmt.Errorf("observability: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = observability.Shutdown(ctx, tp)
	}()

	run := cliexec.New()
	local := localgit.NewInspector(run)
	dash := dashboard.New(forge.NewGitHub(run), forge.NewGitLab(run), local)
	prune, err := pruneagent.New(config.AgentsDir(), run)
	if err != nil {
		return fmt.Errorf("prune agent: %w", err)
	}
	handler := server.NewMux(server.Options{
		Addr:        *addr,
		Projects:    doc.Projects,
		Local:       doc.Local,
		Upstream:    doc.Upstream,
		Dash:        dash,
		Triage:      triage.New(doc.EffectiveLLM()),
		Prune:       prune,
		PollSeconds: doc.EffectivePollSeconds(),
	})
	log.Printf("gitboard: %s (%d projects, config %s, poll %ds, heads cache %ds, merged cache %ds)",
		*addr, len(doc.Projects), path, doc.EffectivePollSeconds(),
		doc.EffectiveHeadsSeconds(), doc.EffectiveMergedSeconds())
	log.Printf("gitboard: uses gh and glab; local roots=%d; AI triage via llm in config; agents %s",
		len(doc.Local.Roots), config.AgentsDir())
	return http.ListenAndServe(*addr, handler)
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	configPath := fs.String("config", config.DefaultPath(), "config file path")
	hostFilter := fs.String("host", "", "limit discovery to github or gitlab")
	addPath := fs.String("add", "", "non-interactive: add owner/repo (requires -host)")
	removeID := fs.String("remove", "", "non-interactive: remove project by id")
	dryRun := fs.Bool("dry-run", false, "discover and print candidates without writing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := *configPath
	if !fileExists(path) {
		if _, err := config.Init(path); err != nil {
			return err
		}
		fmt.Printf("created %s\n", path)
	}
	doc, err := config.Load(path)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*removeID) != "" {
		updated, found := syncproj.RemoveProject(doc.Projects, *removeID)
		if !found {
			return fmt.Errorf("project id %q not found", *removeID)
		}
		doc.Projects = updated
		if *dryRun {
			fmt.Printf("dry-run: would remove %s (%d projects left)\n", *removeID, len(doc.Projects))
			return nil
		}
		if err := config.Save(path, doc); err != nil {
			return err
		}
		fmt.Printf("removed %s (%d projects)\n", *removeID, len(doc.Projects))
		return nil
	}

	if strings.TrimSpace(*addPath) != "" {
		host := config.Host(strings.ToLower(strings.TrimSpace(*hostFilter)))
		if host != config.HostGitHub && host != config.HostGitLab {
			return fmt.Errorf("--add requires --host github|gitlab")
		}
		updated, err := syncproj.AddProject(doc.Projects, host, *addPath)
		if err != nil {
			return err
		}
		doc.Projects = updated
		if *dryRun {
			fmt.Printf("dry-run: would add %s %s (%d projects)\n", host, *addPath, len(doc.Projects))
			return nil
		}
		if err := config.Save(path, doc); err != nil {
			return err
		}
		fmt.Printf("tracked %s %s (%d projects)\n", host, *addPath, len(doc.Projects))
		return nil
	}

	if !doc.Sync.HasSyncSources() {
		if *dryRun {
			return fmt.Errorf("no sync sources: set sync.github.orgs / sync.gitlab.groups in %s (or run sync without --dry-run)", path)
		}
		fmt.Println("No sync sources configured.")
		orgs, groups, err := syncproj.PromptSyncSources()
		if err != nil {
			return err
		}
		doc.Sync.GitHub.Orgs = orgs
		doc.Sync.GitLab.Groups = groups
		if !doc.Sync.HasSyncSources() {
			return fmt.Errorf("need at least one github org or gitlab group")
		}
		if err := config.Save(path, doc); err != nil {
			return err
		}
		fmt.Printf("saved sync sources to %s\n", path)
	}

	run := cliexec.New()
	run.Timeout = 120 * time.Second
	lister := syncproj.ForgeLister{
		GitHub: forge.NewGitHub(run),
		GitLab: forge.NewGitLab(run),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cands, err := syncproj.Discover(ctx, lister, doc, *hostFilter)
	if err != nil {
		return err
	}
	if len(cands) == 0 {
		return fmt.Errorf("no repositories found for configured sync sources")
	}

	fmt.Printf("Candidates: %d repositories\n", len(cands))
	if *dryRun {
		syncproj.FormatCandidates(os.Stdout, cands)
		fmt.Println("dry-run: not writing config")
		return nil
	}

	selected, err := syncproj.SelectInteractive(cands)
	if err != nil {
		return err
	}
	projects, err := syncproj.ApplySelection(cands, selected, doc.Projects)
	if err != nil {
		return err
	}
	doc.Projects = projects
	if err := config.Save(path, doc); err != nil {
		return err
	}
	fmt.Printf("wrote %d projects to %s\n", len(doc.Projects), path)
	return nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
