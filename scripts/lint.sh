#!/usr/bin/env bash
# Lint + format gate for the Go backend and the deploy scripts.
#
# Why a script instead of just `golangci-lint run`: golangci-lint 2.14's built-in
# gofumpt formatter is unreliable (see .golangci.yml), so formatting is checked
# with the standalone binaries, which are deterministic and agree with `gopls`.
#
# Every tool is required: a missing formatter used to print an empty diff and
# report success, which is worse than failing loudly.
#
# Usage:  ./scripts/lint.sh          (check)
#         ./scripts/lint.sh --fix    (apply formatting fixes first)
set -uo pipefail
cd "$(dirname "$(readlink -f "$0")")/.." || exit 1

GOBIN="${GOBIN:-$HOME/.local/share/go/bin}"
fail=0
missing=0

step() { printf '\n\033[1;36m== %s\033[0m\n' "$1"; }
ok()   { printf '\033[0;32m%s\033[0m\n' "$1"; }
bad()  { printf '\033[0;31m%s\033[0m\n' "$1"; fail=1; }
need() { command -v "$1" >/dev/null 2>&1 || [ -x "$1" ]; }

step "Tool preflight"
for tool in gofumpt gci staticcheck golangci-lint; do
  if [ -x "$GOBIN/$tool" ]; then
    continue
  fi
  bad "missing $GOBIN/$tool"
  missing=1
done
for tool in go shellcheck; do
  if need "$tool"; then
    continue
  fi
  bad "missing $tool on PATH"
  missing=1
done
if [ "$missing" -ne 0 ]; then
  bad "ABORTING: install the missing tools (see AGENTS.md → Lint) and re-run"
  exit 1
fi
ok "all tools present"

if [ "${1:-}" = "--fix" ]; then
  step "Formatting (gofumpt + gci)"
  "$GOBIN/gofumpt" -w .
  "$GOBIN/gci" write -s standard -s default -s 'Prefix(webgx)' .
fi

step "Formatting check"
unformatted=$("$GOBIN/gofumpt" -l . 2>/dev/null | grep -v '^$' || true)
if [ -n "$unformatted" ]; then
  bad "gofumpt wants to rewrite:"
  echo "$unformatted"
  echo "run: ./scripts/lint.sh --fix"
else
  ok "gofumpt: clean"
fi

unsorted=$("$GOBIN/gci" diff -s standard -s default -s 'Prefix(webgx)' . 2>/dev/null || true)
if [ -n "$unsorted" ]; then
  bad "gci wants to reorder imports (run ./scripts/lint.sh --fix)"
else
  ok "gci: clean"
fi

step "go build"
go build ./... || bad "build failed"

step "go vet"
go vet ./... || bad "vet failed"

step "staticcheck"
"$GOBIN/staticcheck" ./... || bad "staticcheck failed"

step "golangci-lint"
"$GOBIN/golangci-lint" run ./... || bad "golangci-lint failed"

step "shellcheck (deploy scripts)"
shellcheck -x main.sh scripts/lint.sh || bad "shellcheck failed"

echo
if [ "$fail" -ne 0 ]; then
  bad "LINT FAILED"
  exit 1
fi
ok "ALL CHECKS PASSED"
