#!/usr/bin/env bash
# 单机安全换节点二进制：gzip → 4MB 分片上传 → 远端解压校验 → 容器内原子替换 → 重启
set -euo pipefail
HOST="$1"
BIN="/Users/night/Workspaces/security_workspace/_tools/ScopeSentry-Scan/ss-node-new"
LOCAL_SHA=$(shasum -a 256 "$BIN" | cut -d" " -f1)
echo "[$HOST] local sha=${LOCAL_SHA:0:16}"

# 164MB 直传要几分钟且易被 reset；gzip 后约 50MB，分片更少更快（同 release.sh 的做法）
GZ=$(mktemp -t node-XXXXXX.gz)
gzip -9 -c "$BIN" > "$GZ"
CH=$(mktemp -d); split -b 4m "$GZ" "$CH/part_"
N=$(ls "$CH"/part_* | wc -l | tr -d " ")
echo "[$HOST] 压缩后 $(du -h "$GZ" | cut -f1) / $N 片"
ssh -o BatchMode=yes "$HOST" "rm -rf /tmp/nc && mkdir -p /tmp/nc"
ok=0; START=$(date +%s)
for f in "$CH"/part_*; do
  b=$(basename "$f")
  for i in 1 2 3 4; do
    scp -q -o BatchMode=yes -o ConnectTimeout=10 "$f" "$HOST:/tmp/nc/$b" 2>/dev/null && { ok=$((ok+1)); break; }
  done
done
echo "[$HOST] 上传 $ok/$N 片，耗时 $(( $(date +%s) - START ))s"
rm -rf "$CH" "$GZ"
[ "$ok" -eq "$N" ] || { echo "incomplete"; exit 1; }
R=$(ssh -o BatchMode=yes "$HOST" "cat /tmp/nc/part_* | gunzip > /tmp/ss-node-new && rm -rf /tmp/nc && chmod +x /tmp/ss-node-new && sha256sum /tmp/ss-node-new" | cut -d" " -f1)
[ "${R:0:16}" = "${LOCAL_SHA:0:16}" ] || { echo "[$HOST] sha mismatch: $R"; exit 1; }
echo "[$HOST] sha ok"
scp -q -o BatchMode=yes /tmp/swap-node-bin.sh "$HOST:/tmp/swap.sh"
ssh -o BatchMode=yes "$HOST" "bash /tmp/swap.sh"
