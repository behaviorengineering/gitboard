// Package config loads and saves the gitboard config file: path discovery
// (GITBOARD_CONFIG, XDG user config, working directory), YAML encoding,
// init, and environment overlays. The document types are pkg/boardconfig.
package config

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// DefaultLLMModel is the Polypus/cf_local model used when config omits llm.model.
const DefaultLLMModel = "cf_local/@cf/qwen/qwen3-30b-a3b-fp8"

// Failure-dump retention defaults. defaultExample repeats them; a test keeps
// the two in step.
const (
	DefaultDumpMaxAgeHours = 48
	DefaultDumpMaxFiles    = 20
)

// defaultServiceName is the OpenInference service name when the file has none.
const defaultServiceName = "gitboard"

// defaultExample is the template written by init.
const defaultExample = `llm:
  base_url: http://127.0.0.1:1320/v1
  model: cf_local/@cf/qwen/qwen3-30b-a3b-fp8
  api_key: ""

# OpenInference tracing + error-only dumps for agent/LLM work.
openinference:
  enabled: true
  endpoint: ""
  service_name: gitboard
  failure_dump:
    enabled: true
    dir: ""
    max_age_hours: 48
    max_files: 20

ui:
  poll_seconds: 30
  # Optional path.Match patterns; empty / omitted = show all.
  # Hides matching names from the Branches column only (prune data unchanged).
  # hide_branches:
  #   - majordomo-context/*
  #   - dependabot/*

# Server-side forge cache TTLs (0 = always refetch that slice).
upstream:
  heads_seconds: 120
  merged_seconds: 600

# Optional: scan these trees for local checkouts (incl. git worktrees).
# Match is via origin remote → project host/path. Override per project with local_path.
# fetch_seconds TTL-gates git fetch origin before ↑/↓ sync (0 = always fetch).
# scan_seconds TTL-gates root ScanRoots (0 = always rescan; default 300).
local:
  roots: []
  fetch_seconds: 120
  scan_seconds: 300

sync:
  github:
    orgs: []
  gitlab:
    groups: []
  azuredevops:
    orgs: []
    projects: []
  bitbucket:
    workspaces: []

# Named board views (subsets of projects). Empty / omitted → implicit "default"
# view containing every tracked project. Membership may overlap across views.
# views:
#   - id: work
#     label: Work
#     projects: [demo]

# Curated by: gitboard sync
# Optional per project: local_path: ~/code/my-clone
projects: []
`

// DefaultPollSeconds is used when ui.poll_seconds is unset (zero means use default on load).
const DefaultPollSeconds = 30

// dir returns ~/.config/gitboard (or $XDG_CONFIG_HOME/gitboard).
func dir() string {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "gitboard")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "gitboard")
	}
	return filepath.Join(home, ".config", "gitboard")
}

// DefaultPath resolves the config file location.
// Order: GITBOARD_CONFIG, GITBOARD_PROJECTS (legacy), user config.yaml if present,
// cwd config.yaml, cwd projects.yaml (legacy), else the user config.yaml path
// (may not exist yet). The agent guide in internal/cli lists the same order.
func DefaultPath() string {
	if v := envValue("GITBOARD_CONFIG", "GITBOARD_PROJECTS"); v != "" {
		return expandHome(v)
	}
	userPath := filepath.Join(dir(), "config.yaml")
	if Exists(userPath) {
		return userPath
	}
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil && cwd != "" {
		for _, name := range []string{"config.yaml", "projects.yaml"} {
			p := filepath.Join(cwd, name)
			if Exists(p) {
				return p
			}
		}
	}
	return userPath
}

// expandHome expands a leading ~ through localgit.ExpandPath and keeps the
// value as given when that fails, so the later read reports the bad path.
func expandHome(p string) string {
	expanded, err := localgit.ExpandPath(p)
	if err != nil {
		return p
	}
	return expanded
}

// Exists reports whether path is an existing regular file (not a directory).
func Exists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// Load reads a config YAML file (full document or legacy projects-only) and
// normalises it with boardconfig.File.Normalize.
func Load(path string) (boardconfig.File, error) {
	raw, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec // G304: the operator chooses the config path (flag, env or discovery).
	if err != nil {
		return boardconfig.File{}, fmt.Errorf("read config: %w", err)
	}
	var doc boardconfig.File
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return boardconfig.File{}, fmt.Errorf("parse config: %w", err)
	}
	if err := doc.Normalize(); err != nil {
		return boardconfig.File{}, err
	}
	if err := checkPollEnv(); err != nil {
		return boardconfig.File{}, err
	}
	return doc, nil
}

// checkPollEnv rejects a GITBOARD_POLL_SECONDS that EffectivePollSeconds
// would otherwise ignore, so a typo is reported instead of falling back to
// the default.
func checkPollEnv() error {
	v := strings.TrimSpace(os.Getenv(envPollSeconds))
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("%s=%q: %w", envPollSeconds, v, err)
	}
	if n < 0 {
		return fmt.Errorf("%s=%d: must be 0 or more", envPollSeconds, n)
	}
	return nil
}

// Save normalises doc and writes the config file with mode 0600.
func Save(path string, doc boardconfig.File) error {
	if err := doc.Normalize(); err != nil {
		return err
	}
	raw, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Init writes the default config if path does not exist.
func Init(path string) (created bool, err error) {
	if Exists(path) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, []byte(defaultExample), 0o600); err != nil {
		return false, fmt.Errorf("write config: %w", err)
	}
	return true, nil
}

// EffectiveLLM merges file LLM settings with environment overrides.
func EffectiveLLM(f boardconfig.File) boardconfig.LLM {
	base := cmp.Or(
		envValue("GITBOARD_LLM_BASE_URL", "POLYPUS_BASE_URL"),
		strings.TrimSpace(f.LLM.BaseURL),
	)
	if base != "" && !strings.HasSuffix(base, "/v1") {
		base = strings.TrimRight(base, "/") + "/v1"
	}
	return boardconfig.LLM{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey: cmp.Or(
			envValue("GITBOARD_LLM_API_KEY", "OPENAI_API_KEY"),
			strings.TrimSpace(f.LLM.APIKey),
		),
		Model: cmp.Or(
			envValue("GITBOARD_LLM_MODEL"),
			strings.TrimSpace(f.LLM.Model),
			DefaultLLMModel,
		),
	}
}

// envValue returns the first of the named environment variables that is set
// to a non-blank value, trimmed, or "".
func envValue(names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

// envPollSeconds overrides ui.poll_seconds. Load rejects an invalid value.
const envPollSeconds = "GITBOARD_POLL_SECONDS"

// EffectivePollSeconds returns the UI poll interval in seconds.
// Env GITBOARD_POLL_SECONDS overrides the file. Omitted file value defaults to 30.
// Explicit 0 disables polling.
func EffectivePollSeconds(f boardconfig.File) int {
	if v := strings.TrimSpace(os.Getenv(envPollSeconds)); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 {
			return n
		}
	}
	if f.UI.PollSeconds == nil {
		return DefaultPollSeconds
	}
	if *f.UI.PollSeconds < 0 {
		return DefaultPollSeconds
	}
	return *f.UI.PollSeconds
}

// AgentsDir is where agentsession stores sessions (<config dir>/agents).
func AgentsDir() string {
	return filepath.Join(dir(), "agents")
}

// EffectiveOpenInference merges file settings with env for tracing and
// failure dumps. The endpoint is GITBOARD_OPENINFERENCE_ENDPOINT, then
// POLYPUS_OTLP_ENDPOINT, then openinference.endpoint.
func EffectiveOpenInference(f boardconfig.File) boardconfig.OpenInference {
	enabled := true
	if f.OpenInference.Enabled != nil {
		enabled = *f.OpenInference.Enabled
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_OPENINFERENCE_ENABLED")); v != "" {
		enabled = v == "1" || strings.EqualFold(v, "true")
	}
	dumpEnabled := true
	if f.OpenInference.FailureDump.Enabled != nil {
		dumpEnabled = *f.OpenInference.FailureDump.Enabled
	}
	en := enabled
	den := dumpEnabled
	dumpDir := strings.TrimSpace(f.OpenInference.FailureDump.Dir)
	if dumpDir == "" {
		dumpDir = filepath.Join(dir(), "logs", "inference-failures")
	}
	maxAge := f.OpenInference.FailureDump.MaxAgeHours
	if maxAge <= 0 {
		maxAge = DefaultDumpMaxAgeHours
	}
	maxFiles := f.OpenInference.FailureDump.MaxFiles
	if maxFiles <= 0 {
		maxFiles = DefaultDumpMaxFiles
	}
	return boardconfig.OpenInference{
		Enabled: &en,
		// OTEL_EXPORTER_OTLP_* apply after these, inside olly, when all are empty.
		Endpoint: cmp.Or(
			envValue("GITBOARD_OPENINFERENCE_ENDPOINT", "POLYPUS_OTLP_ENDPOINT"),
			strings.TrimSpace(f.OpenInference.Endpoint),
		),
		ServiceName: cmp.Or(strings.TrimSpace(f.OpenInference.ServiceName), defaultServiceName),
		FailureDump: boardconfig.FailureDump{
			Enabled:     &den,
			Dir:         dumpDir,
			MaxAgeHours: maxAge,
			MaxFiles:    maxFiles,
		},
	}
}
