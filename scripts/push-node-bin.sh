#!/usr/bin/env bash
# 单机安全换节点二进制：分片上传 → sha256 校验 → 远端脚本原子替换 → 重启
set -euo pipefail
HOST="$1"
BIN="/Users/night/Workspaces/security_workspace/_tools/ScopeSentry-Scan/ss-node-new"
LOCAL_SHA=$(shasum -a 256 "$BIN" | cut -d" " -f1)
echo "[$HOST] local sha=${LOCAL_SHA:0:16}"
CH=$(mktemp -d); split -b 4m "$BIN" "$CH/part_" >/dev/null
N=$(ls "$CH"/part_* | wc -l | tr -d " ")
ssh -o BatchMode=yes "$HOST" "rm -rf /tmp/nc && mkdir -p /tmp/nc"
ok=0
for f in "$CH"/part_*; do
  b=$(basename "$f")
  for i in 1 2 3 4; do
    scp -q -o BatchMode=yes -o ConnectTimeout=10 "$f" "$HOST:/tmp/nc/$b" 2>/dev/null && { ok=$((ok+1)); break; }
  done
done
rm -rf "$CH"
echo "[$HOST] chunks $ok/$N"
[ "$ok" -eq "$N" ] || { echo "incomplete"; exit 1; }
R=$(ssh -o BatchMode=yes "$HOST" "cat /tmp/nc/part_* > /tmp/ss-node-new && rm -rf /tmp/nc && chmod +x /tmp/ss-node-new && sha256sum /tmp/ss-node-new" | cut -d" " -f1)
[ "${R:0:16}" = "${LOCAL_SHA:0:16}" ] || { echo "[$HOST] sha mismatch"; exit 1; }
echo "[$HOST] sha ok"
scp -q -o BatchMode=yes /tmp/swap-node-bin.sh "$HOST:/tmp/swap.sh"
ssh -o BatchMode=yes "$HOST" "bash /tmp/swap.sh"
