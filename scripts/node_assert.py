#!/usr/bin/env python3
"""校验服务端返回的节点列表：全部在线且版本一致。

用法: echo '[{"name":..,"version":..,"state":"1"}]' | python3 scripts/node_assert.py <expected_version>
退出码 0=通过，1=有问题。
"""
import json
import sys


def main() -> int:
    want = sys.argv[1] if len(sys.argv) > 1 else ""
    raw = sys.stdin.read().strip()
    if not raw:
        print("   ! 没有拿到节点列表")
        return 1
    try:
        nodes = json.loads(raw)
    except Exception:
        print("   ! 节点列表不是合法 JSON")
        return 1
    if not nodes:
        print("   x 服务端看不到任何节点")
        return 1
    offline = [n.get("name") for n in nodes if str(n.get("state")) != "1"]
    mismatch = [(n.get("name"), n.get("version")) for n in nodes if want and n.get("version") != want]
    print("   节点 %d 台，在线 %d" % (len(nodes), len(nodes) - len(offline)))
    if offline:
        print("   ! 离线: " + ", ".join(offline))
    if mismatch:
        print("   x 版本不一致（期望 %s）: %s" % (want, ", ".join("%s=%s" % (n, v) for n, v in mismatch)))
        return 1
    print("   ok 全部节点在线且版本一致")
    return 0


if __name__ == "__main__":
    sys.exit(main())
