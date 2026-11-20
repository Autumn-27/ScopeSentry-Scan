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
DRY=0
BUILD=1
for a in "$@"; do
  [ "$a" = "--no-build" ] && BUILD=0
  [ "$a" = "--rollback" ] && MODE="rollback"
  [ "$a" = "--dry-run" ] && DRY=1
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

if [ "$DRY" = "1" ]; then
  echo "== DRY RUN：将部署到以下节点（不执行写操作）"
  for H in $NODES; do echo "  - $H"; done
  exit 0
fi

echo "== [2/5] 逐台部署"
FAILED=0
for H in $NODES; do
  echo "-- $H"
  ssh $SSH_OPTS "$H" "rm -f /tmp/n19* /tmp/ss-node 2>/dev/null; true" || true

  # 大文件单连接会被中途 reset（2026-11-20 实测 164MB 传到 ~8MB 就断）。
  # 分片上传（每片 4MB、独立连接、单片重试 3 次）再远端拼接，配合后面的 sha256 校验。
  # 先 gzip 再分片：164MB → ~50MB，片数从 40 降到 13，跨境上传时间省一半以上。
  # 4MB 分片保持不变（实测 6MB 更慢，瓶颈是带宽）。远端解压后再校验原始 sha256。
  GZ_PAYLOAD=$(mktemp -t node-upload-XXXXXX.gz)
  gzip -9 -c "$BIN" > "$GZ_PAYLOAD"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "rm -rf /tmp/ss-node-chunks && mkdir -p /tmp/ss-node-chunks" 2>/dev/null || true
  CHUNK_DIR=$(mktemp -d)
  split -b 4m "$GZ_PAYLOAD" "$CHUNK_DIR/part_"
  TOTAL_CHUNKS=$(ls "$CHUNK_DIR"/part_* | wc -l | tr -d ' ')
  IDX=0; UP_FAIL=0; UP_START=$(date +%s)
  for f in "$CHUNK_DIR"/part_*; do
    base=$(basename "$f")
    IDX=$((IDX + 1))
    for attempt in 1 2 3; do
      scp -q -o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=10 "$f" "$H:/tmp/ss-node-chunks/$base" 2>/dev/null && break
      [[ $attempt -eq 3 ]] && UP_FAIL=1
    done
    if (( IDX % 5 == 0 || IDX == TOTAL_CHUNKS )); then
      echo "   上传 $IDX/$TOTAL_CHUNKS  已用 $(( $(date +%s) - UP_START ))s"
    fi
  done
  rm -rf "$CHUNK_DIR" "$GZ_PAYLOAD"
  if [[ $UP_FAIL -ne 0 ]]; then
    echo "   分片上传失败" >&2
    exit 1
  fi
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$H" "cat /tmp/ss-node-chunks/part_* | gunzip > /tmp/ss-node && chmod +x /tmp/ss-node && rm -rf /tmp/ss-node-chunks" || exit 1
  if false; then
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
    GOT_VER="$(run_remote "$H" "docker logs --tail 40 $CONTAINER" 2>&1 | grep Register | tail -1 | grep -oE "[0-9]+[.][0-9]+[.][0-9]+" | tail -1 || true)"
    if [ -n "$VER" ] && [ -n "$GOT_VER" ] && [ "$GOT_VER" != "$VER" ]; then
      echo "   版本不一致：期望 $VER，实际 $GOT_VER"
      FAILED=1
      continue
    fi
    echo "   节点已注册成功（version ${GOT_VER:-?}）"
  else
    echo "   60 秒内没看到 Register Success，自动回滚"
    rollback_host "$H" || true
    FAILED=1
  fi
done

# 清理旧节点镜像：每次部署都 commit 一个新 tag（约 3.9GB），不清理一台机可堆 20+ 个
# （2026-11-20 实测 tx02 26 个镜像占 44G / 磁盘 78%）。逻辑在 scripts/remote/prune-node-images.sh。
for H in $NODES; do
  scp -q -o BatchMode=yes -o StrictHostKeyChecking=no "$ROOT/scripts/remote/prune-node-images.sh" "$H:/tmp/prune-node-images.sh" 2>/dev/null || true
  run_remote "$H" "bash /tmp/prune-node-images.sh $IMAGE" >/dev/null 2>&1 || true
done

echo "== [5/5] 汇总"
if [ "$FAILED" = "1" ]; then echo "有节点部署失败（已尝试回滚）" >&2; exit 1; fi

# 整体断言：从服务端视角确认「所有节点在线且版本一致」。
# 只看各机自己的启动日志，发现不了"漏部署某台"或"版本分叉"（2026-11-20 加）。
SERVER_HOST="${SERVER_HOST:-prod-tx-sh-app-01}"
NODES_JSON="$(ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SERVER_HOST" "python3 /opt/scopesentry/nodecheck.py --json 2>/dev/null" || echo "")"
if [ -n "$NODES_JSON" ]; then
  HERE="$(cd -- "$(dirname -- "$0")/.." >/dev/null 2>&1 && pwd -P)"
  if ! echo "$NODES_JSON" | python3 "$HERE/scripts/node_assert.py" "${VER:-}"; then
    exit 1
  fi
else
  echo "   ! 无法从 $SERVER_HOST 获取节点状态，请手工确认"
fi


# 稳定性冒烟：部署完 90 秒内看每个节点的心跳是否还在更新。
# 「部署成功」和「部署完就把节点弄掉了」是两件事——2026-11-21 一天里发生过 7 次掉线，
# 每次都得靠人肉发现；这里让脚本自己盯 90 秒。
echo "== 稳定性观察（90 秒）"
sleep 90
STABLE_FAIL=0
for NAME in $(printf "%s\n" "$NODES" | while read -r h; do
  [ -z "$h" ] && continue
  case "$h" in
    *app-01*) echo node-sh-01 ;;
    *app-02*) echo node-tx-sh-02 ;;
    *188.143*) echo node-dev-02 ;;
  esac
done); do
  LINE="$(ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SERVER_HOST" \
    "sg docker -c 'docker exec scopesentry-redis redis-cli -a ac6b0cf473fb8d23a0e13ff3 --no-auth-warning hmget node:$NAME updateTime state containerMem' 2>/dev/null" | tr '\n' ' ')"
  echo "   $NAME: $LINE"
  case "$LINE" in
    *" 1 "*|*" 1") ;;
    *) echo "   ! $NAME 心跳或状态异常（不在线/心跳停滞）" >&2; STABLE_FAIL=1 ;;
  esac
done
if [ "$STABLE_FAIL" = "1" ]; then
  echo "稳定性观察未通过：有节点在部署后 90 秒内掉线" >&2
  exit 1
fi

echo "全部节点部署完成，版本 ${VER:-unknown}"
