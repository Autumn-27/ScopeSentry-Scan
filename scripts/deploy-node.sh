#!/usr/bin/env bash
# 一键部署扫描节点：构建 → 哈希校验 → 远端清理 → 上传 → 换镜像 → 重启 → 验证
#
# 用法：
#   scripts/deploy-node.sh                # 构建并部署到默认节点列表
#   scripts/deploy-node.sh --no-build     # 复用已有二进制
#   NODES="prod-tx-sh-app-01" scripts/deploy-node.sh
#
# 为什么需要（2026-11-20 三次事故）：远端 /tmp 是 tmpfs，历次上传的大二进制会塞满它，
# scp 静默截断后 docker cp 得到半个文件，节点版本号纹丝不动却不报错。
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "$0")/.." >/dev/null 2>&1 && pwd -P)"
BIN="$ROOT/ss-node-deploy"
NODES="${NODES:-prod-tx-sh-app-01 prod-tx-sh-app-02 ubuntu@43.142.188.143}"
CONTAINER="${NODE_CONTAINER:-scopesentry-scan}"
IMAGE="${IMAGE_PREFIX:-autumn27/scopesentry-scan}"

BUILD=1
for a in "$@"; do [ "$a" = "--no-build" ] && BUILD=0; done

if [ "$BUILD" = "1" ]; then
  echo "== [1/5] 构建 linux/amd64 节点二进制"
  docker run --rm -v "$ROOT":/app -w /app -e GOPROXY=https://goproxy.cn,direct \
    -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.24 \
    go build -trimpath -ldflags "-s -w" -o /app/ss-node-deploy ./cmd/ScopeSentry/main.go \
    2>&1 | grep -v downloading || true
else
  echo "== [1/5] 跳过构建，复用 $BIN"
fi
[ -f "$BIN" ] || { echo "构建产物缺失: $BIN" >&2; exit 1; }

SHA="$(shasum -a 256 "$BIN" | cut -d" " -f1)"
SIZE="$(wc -c < "$BIN" | tr -d " ")"
# 版本从源码里的常量读，比从二进制 strings 抠更可靠（macOS strings 不可用时会失败）
VER="$(grep -oE 'global[.]VERSION = "[0-9.]+"' internal/config/config.go | grep -oE '[0-9]+[.][0-9]+[.][0-9]+' | head -1 || true)"
echo "   sha=${SHA:0:16} size=$SIZE version=${VER:-unknown}"

echo "== [2/5] 逐台部署"
FAILED=0
for H in $NODES; do
  echo "-- $H"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 "$H" \
    "rm -f /tmp/n19* /tmp/ss-node* 2>/dev/null; true" || true
  if ! scp -o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -q "$BIN" "$H:/tmp/ss-node"; then
    echo "   上传失败"; FAILED=1; continue
  fi
  RSHA="$(ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "sha256sum /tmp/ss-node | cut -d' ' -f1")"
  if [ "$RSHA" != "$SHA" ]; then
    echo "   哈希不一致（远端 ${RSHA:0:16}），上传被截断，跳过该机"; FAILED=1; continue
  fi
  echo "   上传完整（sha256 一致）"
  WRAP=""
  if ! ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "docker ps >/dev/null 2>&1"; then
    WRAP="sg docker -c "
  fi
  CMD="docker cp /tmp/ss-node $CONTAINER:/apps/ScopeSentry-Scan && docker commit $CONTAINER $IMAGE:${VER:-latest} >/dev/null && docker restart $CONTAINER >/dev/null"
  if [ -n "$WRAP" ]; then
    ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "${WRAP}"$CMD"" || { echo "   换镜像失败"; FAILED=1; continue; }
  else
    ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "$CMD" || { echo "   换镜像失败"; FAILED=1; continue; }
  fi
  echo "   已换镜像并重启（tag ${VER:-latest}）"
  OK=0
  for _ in 1 2 3 4 5 6 7 8 9 10 11 12; do
    sleep 5
    if ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "docker logs --tail 40 $CONTAINER 2>&1 | grep -q 'Register Success'" 2>/dev/null; then OK=1; break; fi
    if [ -n "$WRAP" ] && ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "sg docker -c \"docker logs --tail 40 $CONTAINER\" 2>&1 | grep -q 'Register Success'" 2>/dev/null; then OK=1; break; fi
  done
  if [ "$OK" = "1" ]; then echo "   节点已注册成功"; else echo "   60 秒内没看到 Register Success，请检查节点日志"; FAILED=1; fi
done

echo "== [5/5] 汇总"
if [ "$FAILED" = "1" ]; then echo "有节点部署失败" >&2; exit 1; fi
echo "全部节点部署完成，版本 ${VER:-unknown}"
