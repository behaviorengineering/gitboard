#!/usr/bin/env bash
# Smoke tests for auto-patch-path-predicates.sh (no git required).
set -euo pipefail

_script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=auto-patch-path-predicates.sh
source "${_script_dir}/auto-patch-path-predicates.sh"

failures=0

check() {
  local fn="$1" want="$2" path="$3"
  local got="false"
  if "${fn}" "${path}"; then got="true"; fi
  if [[ "${got}" != "${want}" ]]; then
    echo "FAIL: ${fn} ${path}: want ${want}, got ${got}" >&2
    failures=$((failures + 1))
  fi
}

check is_harness_only_path true  "lefthook.yml"
check is_harness_only_path true  "lefthook/lefthook-go-standalone.yml"
check is_harness_only_path true  "scripts/git-hooks/default-branch-refuse.sh"
check is_harness_only_path true  ".claude/skills/foo/SKILL.md"
check is_harness_only_path true  "CLAUDE.md"
check is_harness_only_path true  "pkg/engine/AGENTS.md"
check is_harness_only_path false "engine/foo.go"
check is_harness_only_path false ".gitlab/ci/auto-patch-release.yml"
check is_harness_only_path false "tools/agent-pack"

AUTO_PATCH_EXTRA_HARNESS_PATHS="tools/agent-pack tools/agent-pack/*"
check is_harness_only_path true  "tools/agent-pack"
check is_harness_only_path true  "tools/agent-pack/skills/x.md"
unset AUTO_PATCH_EXTRA_HARNESS_PATHS

check is_packaging_only_path true  "Dockerfile"
check is_packaging_only_path true  "Dockerfile.prod"
check is_packaging_only_path true  "scripts/docker-build.sh"
check is_packaging_only_path true  ".github/workflows/docker-release.yml"
check is_packaging_only_path true  ".gitlab/ci/docker-release.yml"
check is_packaging_only_path false "pkg/foo/foo.go"
check is_packaging_only_path false "deploy/helm/values.yaml"

AUTO_PATCH_EXTRA_PACKAGING_PATHS="deploy/helm/*"
check is_packaging_only_path true  "deploy/helm/values.yaml"
unset AUTO_PATCH_EXTRA_PACKAGING_PATHS

if [[ "${failures}" -gt 0 ]]; then
  echo "${failures} predicate test(s) failed" >&2
  exit 1
fi

echo "auto-patch-decide_test: ok"
