#!/usr/bin/env bash
# 节点镜像清理：保留当前版本、上一版本和 latest，其余全删。
# 节点每次部署 docker commit 一个新 tag（约 3.9GB），不清理一台机可堆 20+ 个
# （2026-11-20 实测 tx02 26 个镜像占 44G / 磁盘 78%）。
set -uo pipefail
IMG="${1:-autumn27/scopesentry-scan}"
shift || true
EXTRA="$*"
run() { if docker ps >/dev/null 2>&1; then bash -c "$1"; else sg docker -c "$1"; fi; }
ALL=$(run "docker images $IMG --format \"{{.Repository}}:{{.Tag}}\"" | grep -v "<none>" || true)
SEMVER=$(echo "$ALL" | sed "s/.*://" | grep -E "^[0-9]+[.][0-9]+[.][0-9]+$" | sort -V || true)
CUR=$(echo "$SEMVER" | tail -1)
PREV=$(echo "$SEMVER" | tail -2 | head -1)
KEEP="$CUR $PREV $EXTRA latest"
echo "keep: $KEEP"
REMOVED=0
for img in $ALL; do
  tag="${img##*:}"
  case " $KEEP " in
    *" $tag "*) ;;
    *) run "docker rmi $img" >/dev/null 2>&1 && REMOVED=$((REMOVED+1)) ;;
  esac
done
run "docker image prune -f" >/dev/null 2>&1 || true
echo "removed: $REMOVED"
df -h / | tail -1