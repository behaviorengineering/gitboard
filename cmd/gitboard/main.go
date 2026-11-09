// Command gitboard serves a local code-change board for GitHub, GitLab,
// Azure DevOps and Bitbucket, bound to localhost unless --allow-non-localhost
// is set.
package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/behaviorengineering/gitboard/internal/cli"
)

// version is set by GoReleaser / go tool task build via -ldflags -X main.version=...
// go install does not apply those ldflags, so reportVersion falls back to BuildInfo.
var version = "dev"

func reportVersion() string {
	moduleVersion := ""
	if bi, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = bi.Main.Version
	}
	return resolveVersion(version, moduleVersion)
}

func resolveVersion(ldflag, moduleVersion string) string {
	if v := strings.TrimSpace(ldflag); v != "" && v != "dev" {
		return v
	}
	if v := strings.TrimSpace(moduleVersion); v != "" && v != "(devel)" {
		return v
	}
	return "dev"
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Execute(ctx, cli.Options{Version: reportVersion()}, os.Args[1:])
	stop()
	os.Exit(code)
}
