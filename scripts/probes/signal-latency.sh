#!/usr/bin/env bash
# Probe: signal -> exit latency on a stuck child should be < 5s.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"
ensure_clean_env

WORKDIR=$(mk_tmpdir)
trap 'rm -rf "$WORKDIR"' EXIT

cat > "$WORKDIR/sleeper.sh" <<'EOF'
#!/usr/bin/env bash
trap 'echo got SIGINT; exit 0' INT
sleep 30
EOF
chmod +x "$WORKDIR/sleeper.sh"

start_ts=$(date +%s)
"$WORKDIR/sleeper.sh" &
pid=$!
sleep 0.5
kill -INT "$pid"
wait "$pid" 2>/dev/null || true
end_ts=$(date +%s)
elapsed=$(( end_ts - start_ts ))

echo "signal-to-exit: ${elapsed}s"
if [ "$elapsed" -le 5 ]; then
  echo "PASS"
  exit 0
fi
echo "FAIL"
exit 1
