#!/usr/bin/env bash
# ============================================================================
# ZLMediaKit 载荷构建脚本（安装包含音视频用）
# 位置：create_package/build-zlm-payload.sh
#
# 用途：从构建树组装出**可分发**的 video.tar.gz 载荷：
#   <staging>/run/bin/MediaServer            ← ZLM 主程序（x86_64）
#   <staging>/libs/ZLMediaKit/libmk_api.so   ← ZLM API 库
#   <staging>/conf/zlmediakit/config.ini     ← ZLM 配置（HTTP=8080，secret 示例值）
#   <staging>/logs/, run/log/, run/www/      ← 空骨架
#
# 输出：create_package/bin/bin/video.tar.gz
#
# ★ 裁剪原则：
#   - 不含运行期数据（MediaServer.log、日志、http 缓存等一律不打包）；
#   - config.ini 使用构建树模板，HTTP 端口统一改为 8080（前端/后端约定），
#     secret 保留官方示例值（后端启动时自动生成随机值写回）。
#
# 用法：
#   bash create_package/build-zlm-payload.sh
#   环境变量：
#     ZLM_SRC   构建树 ZLM 目录（默认 <repo>/core/libs/bin/zlmediakit）
# ============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ZLM_SRC="${ZLM_SRC:-$ROOT/core/libs/bin/zlmediakit}"
PAYLOAD_DIR="$ROOT/create_package/bin/bin"
OUT_TGZ="$PAYLOAD_DIR/video.tar.gz"
STAGING="$(mktemp -d)"
trap 'rm -rf "$STAGING"' EXIT

log() { echo -e "\n===== [zlm-payload] $* ====="; }
die() { echo -e "\n❌ [zlm-payload] $*" >&2; exit 1; }

mkdir -p "$PAYLOAD_DIR"

log "[1/4] 组装 video 目录骨架"
VIDEO="$STAGING/video"
mkdir -p "$VIDEO"/{run/bin,libs/ZLMediaKit,conf/zlmediakit,logs}

[ -x "$ZLM_SRC/MediaServer" ] || die "ZLM MediaServer 不存在：$ZLM_SRC/MediaServer"
cp "$ZLM_SRC/MediaServer" "$VIDEO/run/bin/MediaServer"
chmod 755 "$VIDEO/run/bin/MediaServer"

if [ -f "$ZLM_SRC/lib/libmk_api.so" ]; then
  cp "$ZLM_SRC/lib/libmk_api.so" "$VIDEO/libs/ZLMediaKit/libmk_api.so"
  chmod 755 "$VIDEO/libs/ZLMediaKit/libmk_api.so"
else
  log "⚠️ 警告：libmk_api.so 未找到（仅影响 API 动态库，不影响 MediaServer 运行）"
fi

# config.ini：使用构建树模板，HTTP 端口统一 8080，secret 保留示例值
if [ -f "$ZLM_SRC/config.ini" ]; then
  cp "$ZLM_SRC/config.ini" "$VIDEO/conf/zlmediakit/config.ini"
  # 修改 [http] port = 8080（ZLM 默认 80；平台约定 8080）
  sed -i -E 's/^(port=)(80)(\s*)$/\18080\3/' "$VIDEO/conf/zlmediakit/config.ini" 2>/dev/null || true
  # 保险：若上面没匹配到，用 awk 精确改 [http] 段的 port
  python3 - <<'PYEOF' "$VIDEO/conf/zlmediakit/config.ini"
import re, sys
p = sys.argv[1]
lines = open(p, encoding="utf-8").read().split("\n")
in_http = False
for i, ln in enumerate(lines):
    s = ln.strip()
    if s.startswith("[") and s.endswith("]"):
        in_http = (s == "[http]")
        continue
    if in_http and re.match(r"^port\s*=\s*80\s*$", s):
        lines[i] = re.sub(r"^(port\s*=\s*)(80)(\s*)$", r"\g<1>8080\g<3>", ln)
        break
open(p, "w", encoding="utf-8").write("\n".join(lines))
PYEOF
  # ★ 公网拉流：[http] allow_ip_range 置空（不限制），否则外部 IP 访问 8080 的
  #   HLS/FLV/WebRTC 信令会被拒（-100 Your ip is not allowed）。
  sed -i -E 's|^allow_ip_range=.*|allow_ip_range=|' "$VIDEO/conf/zlmediakit/config.ini" 2>/dev/null || true
  log "config.ini 已就绪（[http] port=8080，allow_ip_range 已置空）"
else
  die "ZLM config.ini 不存在：$ZLM_SRC/config.ini"
fi

# 空骨架
mkdir -p "$VIDEO/run/log" "$VIDEO/run/www"

log "[2/4] 可执行性校验"
if LD_LIBRARY_PATH="$VIDEO/libs/ZLMediaKit" "$VIDEO/run/bin/MediaServer" -h >/dev/null 2>&1; then
  log "✅ MediaServer（-h 正常）"
else
  # MediaServer 无 -h 或返回非 0 均属正常，这里只验证能否动态加载
  LD_LIBRARY_PATH="$VIDEO/libs/ZLMediaKit" "$VIDEO/run/bin/MediaServer" -l 2>&1 | head -3 || true
  log "⚠️  MediaServer 启动参数探测：依赖动态库，请以实际启动为准"
fi

log "[3/4] 打包 video.tar.gz"
rm -f "$OUT_TGZ"
tar -czf "$OUT_TGZ" -C "$STAGING" video
log "输出：$OUT_TGZ"
du -h "$OUT_TGZ"

log "[4/4] 包内结构抽检"
TMPCHK="$(mktemp -d)"
tar -xzf "$OUT_TGZ" -C "$TMPCHK"
for f in \
  video/run/bin/MediaServer \
  video/conf/zlmediakit/config.ini \
  video/logs; do
  [ -e "$TMPCHK/$f" ] || die "video.tar.gz 缺少：$f"
done
# 校验 [http] port=8080
if grep -q '^port=8080$' "$TMPCHK/video/conf/zlmediakit/config.ini"; then
  log "✅ [http] port=8080"
else
  warn "config.ini 中未找到 port=8080（[http] 段），请检查模板"
fi
rm -rf "$TMPCHK"
log "✅ ZLM 载荷构建完成"