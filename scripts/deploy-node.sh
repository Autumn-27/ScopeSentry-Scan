#!/usr/bin/env bash
# 一键部署扫描节点：构建 → 哈希校验 → 备份旧二进制 → 上传 → 换镜像 → 验证 → 失败自动回滚
#
# 用法：
#   scripts/deploy-node.sh                 # 构建并部署到默认节点列表
#   scripts/deploy-node.sh --no-build      # 复用已有 ss-node-deploy 二进制
#   scripts/deploy-node.sh --rollback      # 把所有节点回滚到上一次备份的二进制
#   NODES="prod-tx-sh-app-01" scripts/deploy-node.sh
#
# 2026-11-20 事故背景：远端 /tmp 是 tmpfs，历次上传的大二进制会塞满它，scp 静默截断；
# 截断后 docker cp 得到半个文件，节点版本号纹丝不动却不报错。本脚本用 sha256 拦住，
# 并在部署前把容器内旧二进制备份到 /tmp/ss-node.prev，失败自动回滚。
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "$0")/.." >/dev/null 2>&1 && pwd -P)"
BIN="$ROOT/ss-node-deploy"
NODES="${NODES:-prod-tx-sh-app-01 prod-tx-sh-app-02 ubuntu@43.142.188.143}"
CONTAINER="${NODE_CONTAINER:-scopesentry-scan}"
IMAGE="${IMAGE_PREFIX:-autumn27/scopesentry-scan}"

MODE="deploy"
BUILD=1
for a in "$@"; do
  [ "$a" = "--no-build" ] && BUILD=0
  [ "$a" = "--rollback" ] && MODE="rollback"
done

SSH_OPTS="-o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10"

# 有的机器当前会话不在 docker 组，需要 sg docker 包装
docker_wrap() {
  local h="$1"
  if ssh $SSH_OPTS "$h" "docker ps >/dev/null 2>&1"; then echo ""; else echo "sg docker -c "; fi
}

run_remote() {
  local h="$1"; shift
  local cmd="$1"
  local wrap
  wrap="$(docker_wrap "$h")"
  if [ -n "$wrap" ]; then
    ssh $SSH_OPTS "$h" "${wrap}"$cmd""
  else
    ssh $SSH_OPTS "$h" "$cmd"
  fi
}

wait_register() {
  local h="$1" i
  for i in 1 2 3 4 5 6 7 8 9 10 11 12; do
    sleep 5
    if run_remote "$h" "docker logs --tail 40 $CONTAINER 2>&1 | grep -q 'Register Success'" >/dev/null 2>&1; then
      return 0
    fi
  done
  return 1
}

rollback_host() {
  local h="$1"
  echo "   回滚 $h ..."
  if ! run_remote "$h" "docker exec $CONTAINER test -s /apps/ScopeSentry-Scan.prev 2>/dev/null || docker cp $CONTAINER:/apps/ScopeSentry-Scan /tmp/ss-node.prev >/dev/null 2>&1"; then
    : # 容器内没有 .prev 时，退回宿主机 /tmp/ss-node.prev
  fi
  if run_remote "$h" "docker cp /tmp/ss-node.prev $CONTAINER:/apps/ScopeSentry-Scan && docker restart $CONTAINER >/dev/null"; then
    if wait_register "$h"; then echo "   回滚成功"; return 0; fi
  fi
  echo "   回滚失败，需要人工介入"; return 1
}

if [ "$MODE" = "rollback" ]; then
  echo "== 回滚模式"
  FAILED=0
  for H in $NODES; do
    echo "-- $H"
    rollback_host "$H" || FAILED=1
  done
  [ "$FAILED" = "0" ] && echo "全部回滚完成" || { echo "有节点回滚失败" >&2; exit 1; }
  exit 0
fi

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
VER="$(grep -oE 'global[.]VERSION = "[0-9.]+"' internal/config/config.go | grep -oE '[0-9]+[.][0-9]+[.][0-9]+' | head -1 || true)"
echo "   sha=${SHA:0:16} size=$SIZE version=${VER:-unknown}"

echo "== [2/5] 逐台部署"
FAILED=0
for H in $NODES; do
  echo "-- $H"
  ssh $SSH_OPTS "$H" "rm -f /tmp/n19* /tmp/ss-node 2>/dev/null; true" || true

  if ! scp -o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 -q "$BIN" "$H:/tmp/ss-node"; then
    echo "   上传失败"; FAILED=1; continue
  fi
  RSHA="$(ssh $SSH_OPTS "$H" "sha256sum /tmp/ss-node | cut -d' ' -f1")"
  if [ "$RSHA" != "$SHA" ]; then
    echo "   哈希不一致（远端 ${RSHA:0:16}），上传被截断，跳过该机"; FAILED=1; continue
  fi
  echo "   上传完整（sha256 一致）"

  # 备份容器内当前二进制：回滚只需要把它 cp 回去再 restart
  if run_remote "$H" "docker cp $CONTAINER:/apps/ScopeSentry-Scan /tmp/ss-node.prev"; then
    PREV_SIZE="$(ssh $SSH_OPTS "$H" "wc -c < /tmp/ss-node.prev" 2>/dev/null || echo 0)"
    echo "   已备份旧二进制（$PREV_SIZE 字节）"
  else
    echo "   备份旧二进制失败，继续（回滚能力缺失）"
  fi

  if ! run_remote "$H" "docker cp /tmp/ss-node $CONTAINER:/apps/ScopeSentry-Scan && docker commit $CONTAINER $IMAGE:${VER:-latest} >/dev/null && docker restart $CONTAINER >/dev/null"; then
    echo "   换镜像失败，尝试回滚"; rollback_host "$H" || true; FAILED=1; continue
  fi
  echo "   已换镜像并重启（tag ${VER:-latest}）"

  if wait_register "$H"; then
    echo "   节点已注册成功"
  else
    echo "   60 秒内没看到 Register Success，自动回滚"
    rollback_host "$H" || true
    FAILED=1
  fi
done

echo "== [5/5] 汇总"
if [ "$FAILED" = "1" ]; then echo "有节点部署失败（已尝试回滚）" >&2; exit 1; fi
echo "全部节点部署完成，版本 ${VER:-unknown}"
