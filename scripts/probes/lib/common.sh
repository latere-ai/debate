#!/usr/bin/env bash
# Shared helpers for scripts/probes/*.sh.

require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 2; }
}

ensure_clean_env() {
  unset ANTHROPIC_API_KEY AGON_IN_PROGRESS
}

mk_tmpdir() {
  mktemp -d "${TMPDIR:-/tmp}/agon-probe-XXXXXX"
}

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}
