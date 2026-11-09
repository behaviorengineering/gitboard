package cli

import (
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// cliCallBudget caps one forge or git CLI call made by serve or sync.
const cliCallBudget = 120 * time.Second

// configFlagHelp is the --config help on every command. The agent guide
// lists the full discovery order.
const configFlagHelp = "config file path (default: GITBOARD_CONFIG, then the XDG user config.yaml if it exists, " +
	"then ./config.yaml; run gitboard for the full order)"

// newRunner returns the CLI runner serve and sync share for forge and git calls.
func newRunner() *cliexec.Runner {
	run := cliexec.New()
	run.Timeout = cliCallBudget
	return run
}

// dashboardConfig returns the dashboard wiring serve and sync share: the four
// forge clients on run and the process clock. Callers that read or change
// local checkouts set Local before CreateService.
func dashboardConfig(run cliexec.Exec, now clock.Clock) dashboard.ServiceConfig {
	cfg := dashboard.DefaultServiceConfig(run)
	cfg.Clock = now
	return cfg
}

// loadConfig reads the config file. A file that cannot be read or parsed is
// operator input to fix, so it is CodeInvalid (exit code 2).
func loadConfig(path string) (boardconfig.File, error) {
	doc, err := config.Load(path)
	if err != nil {
		return boardconfig.File{}, gberr.Wrap(err, gberr.CodeInvalid, "cli.loadConfig", "cannot load the config file").
			With("path", path)
	}
	return doc, nil
}
