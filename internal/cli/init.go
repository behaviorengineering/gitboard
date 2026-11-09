package cli

import (
	"github.com/spf13/cobra"

	"github.com/behaviorengineering/gitboard/internal/config"
)

func newInitCmd() *cobra.Command {
	var (
		path   string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the config file if missing (never overwrites)",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p := &printer{w: cmd.OutOrStdout()}
			if dryRun {
				if config.Exists(path) {
					p.printf("dry-run: %s already exists (would not overwrite)\n", path)
				} else {
					p.printf("dry-run: would create %s\n", path)
				}
				return p.err
			}
			created, err := config.Init(path)
			if err != nil {
				return err
			}
			if created {
				p.printf("created %s\n", path)
			} else {
				p.printf("already exists: %s (not overwritten)\n", path)
			}
			p.printf("next: edit llm settings if needed, then: gitboard sync\n")
			return p.err
		},
	}
	cmd.Flags().StringVar(&path, "config", config.DefaultPath(), configFlagHelp)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what init would do without writing")
	return cmd
}
