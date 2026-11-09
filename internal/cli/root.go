// Package cli wires the gitboard command tree (Cobra) to the config, sync and
// dashboard services.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Options carries process-level inputs for the command tree.
type Options struct {
	// Version is the build identity printed by `gitboard version`.
	Version string
	// Interactive reports whether stdin is a terminal. Nil means os.Stdin is checked.
	Interactive func() bool
}

// Execute runs the command tree with args and returns the process exit code
// (see exitCode).
func Execute(ctx context.Context, opts Options, args []string) int {
	root := newRoot(opts)
	root.SetArgs(args)
	return exitCode(root.ExecuteContext(ctx))
}

// newRoot builds the gitboard root command with every subcommand attached.
func newRoot(opts Options) *cobra.Command {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if opts.Interactive == nil {
		opts.Interactive = stdinIsTerminal
	}
	var showVersion bool
	root := &cobra.Command{
		Use:           "gitboard",
		Short:         "Local code-change board for GitHub, GitLab, Azure DevOps and Bitbucket",
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showVersion {
				return printVersion(cmd.OutOrStdout(), opts.Version)
			}
			return agentOperatingGuide(cmd.OutOrStdout(), opts.Version)
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usage(err) })
	root.Flags().BoolVar(&showVersion, "version", false, "print the build version")
	// Legacy alias for `gitboard version`, kept off the help catalog.
	if err := root.Flags().MarkHidden("version"); err != nil {
		panic(err) // The flag is defined on the line above.
	}

	root.AddCommand(
		newVersionCmd(opts.Version),
		newInitCmd(),
		newSyncCmd(opts.Interactive),
		newServeCmd(),
	)
	return root
}

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printVersion(cmd.OutOrStdout(), version)
		},
	}
}

func printVersion(w io.Writer, version string) error {
	_, err := fmt.Fprintf(w, "gitboard %s\n", version)
	return err
}

// agentOperatingGuide is what bare `gitboard` prints. It never starts serve.
func agentOperatingGuide(w io.Writer, version string) error {
	lines := []string{
		"gitboard " + version,
		"Local code-change board for GitHub, GitLab, Azure DevOps and Bitbucket.",
		"The dashboard binds to 127.0.0.1 unless told otherwise.",
		"",
		"Bare invoke prints this guide; it never starts serve.",
		"",
		"Documentation for agents",
		"  AGENTS.md",
		"  README.md",
		"",
		"Read-only",
		"  version, help",
		"  sync --dry-run       list candidate repos, write nothing",
		"  init --dry-run       report whether init would create the config",
		"",
		"Changes the local config file",
		"  init                 create the config if missing (never overwrites)",
		"  sync                 interactive repo selection (needs a terminal)",
		"",
		"Long-running",
		"  serve                dashboard on 127.0.0.1:1325",
		"",
		"Automation (no terminal needed)",
		"  gitboard sync --host github --add owner/repo",
		"  gitboard sync --remove <project-id>",
		"  sync without a terminal and without --add, --remove or --dry-run exits 2",
		"",
		"Config path (first match wins)",
		"  --config <path>",
		"  GITBOARD_CONFIG (legacy: GITBOARD_PROJECTS)",
		"  $XDG_CONFIG_HOME/gitboard/config.yaml (default ~/.config/gitboard/config.yaml), if it exists",
		"  ./config.yaml in the current directory (legacy: ./projects.yaml), if it exists",
		"  otherwise the XDG path, which init and serve create",
		"",
		"Exit codes",
		"  0  success",
		"  1  internal error",
		"  2  invalid input or usage (bad flag, unknown host or project, no terminal)",
		"  3  conflict with the current state",
		"  4  not found",
		"  5  a forge, git or the network is unavailable",
		"  6  cancelled or timed out",
		"",
		"Full command list: gitboard help",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// printer writes operator lines and keeps the first write error, so commands
// can print several lines and return one error at the end.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, args ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, args...)
	}
}

// stdinIsTerminal asks the OS whether stdin is a terminal. A file-mode check is
// not enough: /dev/null is a character device too.
func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
