#!/usr/bin/env sh
# install.sh - one-liner installer for `agon`.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/latere-ai/agon/main/install.sh | sh
#
# Optional env:
#   AGON_VERSION=v0.0.1-rc2   # pin a specific tag (default: latest)
#   AGON_PREFIX=/usr/local    # install dir parent (binary lands at $AGON_PREFIX/bin)
#   AGON_NO_HOOK=1            # skip the `install-hook --scope user` step
#
# Requires: curl, tar. Uses sudo only if the install dir is not writable.

set -eu

REPO=latere-ai/agon
PREFIX=${AGON_PREFIX:-/usr/local}
BINDIR="$PREFIX/bin"
VERSION=${AGON_VERSION:-}

err() { printf 'install: %s\n' "$*" >&2; exit 1; }

require() { command -v "$1" >/dev/null 2>&1 || err "missing: $1"; }
require curl
require tar
require uname

# Detect platform.
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$OS" in
  darwin|linux) ;;
  *) err "unsupported OS: $OS (only darwin / linux)" ;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) err "unsupported arch: $ARCH (only amd64 / arm64)" ;;
esac

# Resolve latest tag if VERSION is unset. /releases/latest returns
# 404 while only pre-releases exist; fall back to /releases (which
# lists all, newest first) in that case.
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' \
    | head -1 || true)
  if [ -z "$VERSION" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases" \
      | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' \
      | head -1)
  fi
  [ -n "$VERSION" ] || err "could not resolve latest tag; set AGON_VERSION="
fi

# Asset filename: agon_<version-without-v>_<os>_<arch>.tar.gz
VNUM=${VERSION#v}
ASSET="agon_${VNUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"

printf 'install: %s -> %s\n' "$VERSION ($OS/$ARCH)" "$BINDIR"

TMP=$(mktemp -d -t agon-install.XXXXXX)
trap 'rm -rf "$TMP"' EXIT

curl -fL --progress-bar "$URL" -o "$TMP/$ASSET" \
  || err "download failed: $URL"

# Optional checksum verification.
if curl -fsSL "https://github.com/$REPO/releases/download/$VERSION/checksums.txt" \
   -o "$TMP/checksums.txt" 2>/dev/null; then
  if command -v shasum >/dev/null 2>&1; then
    (cd "$TMP" && shasum -a 256 -c --ignore-missing checksums.txt >/dev/null) \
      || err "checksum mismatch"
  elif command -v sha256sum >/dev/null 2>&1; then
    (cd "$TMP" && sha256sum -c --ignore-missing checksums.txt >/dev/null) \
      || err "checksum mismatch"
  fi
fi

tar -xzf "$TMP/$ASSET" -C "$TMP"
[ -x "$TMP/agon" ] || err "extracted archive does not contain a 'agon' binary"

# Install. Use sudo only if BINDIR is not writable by us.
mkdir -p "$BINDIR" 2>/dev/null || true
if [ -w "$BINDIR" ] || { [ ! -e "$BINDIR" ] && [ -w "$(dirname "$BINDIR")" ]; }; then
  install "$TMP/agon" "$BINDIR/agon"
else
  printf 'install: %s is not writable, using sudo\n' "$BINDIR"
  sudo install "$TMP/agon" "$BINDIR/agon"
fi

# Install Stop hook unless opted out.
if [ "${AGON_NO_HOOK:-0}" != "1" ]; then
  if "$BINDIR/agon" install-hook --scope user; then
    printf 'install: hook installed\n'
  else
    printf 'install: hook install failed; rerun with: %s install-hook --scope user\n' \
      "$BINDIR/agon"
  fi
fi

printf '\ninstalled %s at %s\n' "$("$BINDIR/agon" --version)" "$BINDIR/agon"
