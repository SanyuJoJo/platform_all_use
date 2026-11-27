#!/usr/bin/env bash
# ============================================================
# 生成升级包（可分发）
# 位置：create_package/upgrade/build_upgrade_script.sh
#
# 用法：
#   bash build_upgrade_script.sh --product backend \
#     --from 1.6.0 --to 1.7.0 --channel prod --out <dir>
#
# 产物（写进 <out>/<name>-<to>/）：
#   - manifest.json           升级清单
#   - backend.tar.gz         后端载荷（可选）
#   - dist.war               前端载荷（可选）
#   - manifest.json.sig      签名（需 ED25519_PRIVATE_KEY，离线）
#
# 部署：把整个 <name>-<to>/ 目录放到 <INSTALL_DIR>/update/ 下即可
# 被 /api/v1/upgrade/status 扫描到，并可用 /api/v1/upgrade/run 应用。
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PRODUCT=""
FROM=""
TO=""
CHANNEL="prod"
OUT=""
MIN_PLATFORM=""
REQUIRES=""

usage() {
  cat <<EOF
用法：$0 [选项]
  --product <backend|frontend|core|all>  要打的载荷
  --from <版本>       升级源版本
  --to <版本>         目标版本
  --channel <prod|test>  通道（默认 prod）
  --min-platform <版本>  最低平台版本（可选，默认同 --from）
  --requires <名称>    依赖说明（可选）
  --out <目录>         输出目录（必填）
  -h, --help
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --product) PRODUCT="$2"; shift 2 ;;
    --from) FROM="$2"; shift 2 ;;
    --to) TO="$2"; shift 2 ;;
    --channel) CHANNEL="$2"; shift 2 ;;
    --min-platform) MIN_PLATFORM="$2"; shift 2 ;;
    --requires) REQUIRES="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "未知参数：$1" >&2; exit 2 ;;
  esac
done

[ -n "$PRODUCT" ] || { echo "--product 必填" >&2; exit 2; }
[ -n "$FROM" ] || { echo "--from 必填" >&2; exit 2; }
[ -n "$TO" ] || { echo "--to 必填" >&2; exit 2; }
[ -n "$OUT" ] || { echo "--out 必填" >&2; exit 2; }
case "$PRODUCT" in backend|frontend|core|all) ;; *) echo "非法 --product：$PRODUCT" >&2; exit 2 ;; esac

PAYLOAD="$ROOT/create_package/bin/bin"
NAME="platform-upgrade"
PKG_DIR="$OUT/$NAME-$TO"
mkdir -p "$PKG_DIR"
rm -f "$PKG_DIR"/*.tar.gz "$PKG_DIR"/dist.war "$PKG_DIR"/manifest.json* 2>/dev/null || true

sha256sum_file() { sha256sum "$1" | awk '{print $1}'; }

BACKEND_ARTIFACT=""
BACKEND_SHA=""
FRONTEND_ARTIFACT=""
FRONTEND_SHA=""

# 后端载荷：复用 build.sh --flat
if [ "$PRODUCT" = "backend" ] || [ "$PRODUCT" = "all" ]; then
  [ -f "$PAYLOAD/backend.tar.gz" ] || { echo "缺少 $PAYLOAD/backend.tar.gz（先跑 build-product.sh）" >&2; exit 1; }
  cp -f "$PAYLOAD/backend.tar.gz" "$PKG_DIR/backend.tar.gz"
  BACKEND_ARTIFACT="backend.tar.gz"
  BACKEND_SHA="$(sha256sum_file "$PKG_DIR/backend.tar.gz")"
fi

# 前端载荷
if [ "$PRODUCT" = "frontend" ] || [ "$PRODUCT" = "all" ]; then
  [ -f "$PAYLOAD/dist.war" ] || { echo "缺少 $PAYLOAD/dist.war（先跑 build-product.sh）" >&2; exit 1; }
  cp -f "$PAYLOAD/dist.war" "$PKG_DIR/dist.war"
  FRONTEND_ARTIFACT="dist.war"
  FRONTEND_SHA="$(sha256sum_file "$PKG_DIR/dist.war")"
fi

# 写 manifest.json
cat > "$PKG_DIR/manifest.json" <<EOF
{
  "schema_version": "1.0",
  "name": "$NAME",
  "from_version": "$FROM",
  "to_version": "$TO",
  "min_platform_version": "${MIN_PLATFORM:-$FROM}",
  "channel": "$CHANNEL",
  "requires": "${REQUIRES:-core}",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "backend_artifact": "$BACKEND_ARTIFACT",
  "frontend_artifact": "$FRONTEND_ARTIFACT",
  "backend_sha256": "$BACKEND_SHA",
  "frontend_sha256": "$FRONTEND_SHA",
  "signature_file": "manifest.json.sig"
}
EOF

# 签名（ed25519，私钥离线；无私钥时跳过并告警）
if [ -n "${ED25519_PRIVATE_KEY:-}" ]; then
  if command -v python3 >/dev/null 2>&1; then
    ED25519_PRIVATE_KEY="$ED25519_PRIVATE_KEY" python3 - "$PKG_DIR/manifest.json" <<'PYEOF'
import os, sys, base64
p = sys.argv[1]
# 兼容两种：16 字节 seed 或 64 字节私钥（PEM/DER 或 base64）
key = os.environ["ED25519_PRIVATE_KEY"].strip()
data = open(p, "rb").read()
try:
    import nacl.signing
    if len(key) <= 88 and not key.startswith("-----"):
        seed = base64.b64decode(key)
        sk = nacl.signing.SigningKey(seed)
    else:
        # PEM 或 64 字节
        if key.startswith("-----"):
            from nacl.signing import SigningKey
            pem = key.encode()
            # 提取 DER
            b64 = "".join(pem.split(b"-----")[2].split())
            der = base64.b64decode(b64)
            sk = SigningKey(der[-64:])
        else:
            raw = base64.b64decode(key)
            sk = nacl.signing.SigningKey(raw[-32:])
    sig = sk.sign(data).signature
    open(p + ".sig", "wb").write(sig)
    print("已生成签名：manifest.json.sig")
except Exception as e:
    print("签名失败（缺 PyNaCl 或私钥格式不支持）：", e)
    sys.exit(1)
PYEOF
  else
    echo "⚠️  未安装 python3，跳过签名（升级会因无法校验而被拒绝）" >&2
  fi
else
  echo "⚠️  未设置 ED25519_PRIVATE_KEY，跳过签名。" >&2
  echo "    不签名时 /api/v1/upgrade/run 仍会执行（由 apply.sh 做 sha 强校验）。" >&2
fi

echo "✅ 升级包已生成：$PKG_DIR"
ls -lh "$PKG_DIR"
echo
echo "部署：把 $PKG_DIR 整个目录放到 <INSTALL_DIR>/update/ 下"
echo "确认：GET /api/v1/upgrade/status 可见 update_available=true"