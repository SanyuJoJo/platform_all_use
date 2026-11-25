#!/usr/bin/env bash
# ============================================================
# nshttps_gm 统一入口 nginx 配置生成脚本
#
# 生成 nshttps.conf：监听 8443，国际 TLS + 国密 NTLS 双协议共存，
# 单向认证（仅服务端证书，不强制客户端证书），反代到 127.0.0.1 内部服务。
#
# 路由：
#   /api/*                -> 后端 REST     ${UPSTREAM}  (/api/v1/live, /api/v1/media)
#   音视频播放流(HLS/FLV/TS) -> ZLMediaKit   127.0.0.1:8080
#   /index/api/, /webrtc/  -> ZLMediaKit   127.0.0.1:8080
#   /                     -> 后端前端托管   ${UPSTREAM}
#
# 证书分工：
#   - 国际 TLS 通道：标准 ECC(P-256) 证书 international.* （标准浏览器可用）
#   - 国密 NTLS 通道：SM2 双证书 server.* + server_enc.* （国密浏览器可用）
#
# 用法：
#   bash gen-config.sh [--port 8443] [--upstream 127.0.0.1:8000] [--zlm 127.0.0.1:8080]
# ============================================================
set -euo pipefail

NSG_HOME="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
NGINX_BIN="$NSG_HOME/tengine-3.1.0/bin/sbin/nginx"
MIME_TYPES="$NSG_HOME/tengine-3.1.0/conf/mime.types"
CERT_DIR="$NSG_HOME/certs"
CONF_DIR="$NSG_HOME/conf"
LOG_DIR="$NSG_HOME/logs"
PID_DIR="$NSG_HOME/pid"
TMP_DIR="$NSG_HOME/tmp"
RUN_DIR="$NSG_HOME/run"
mkdir -p "$CONF_DIR" "$LOG_DIR" "$PID_DIR" "$TMP_DIR" "$RUN_DIR"

# ---------- 参数 ----------
PORT=8443
UPSTREAM="127.0.0.1:8000"
ZLM="127.0.0.1:8080"
while [ $# -gt 0 ]; do
  case "$1" in
    --port)     PORT="$2";     shift 2 ;;
    --upstream) UPSTREAM="$2"; shift 2 ;;
    --zlm)      ZLM="$2";      shift 2 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

[ -f "$CERT_DIR/ca/sm2-root.crt" ] || { echo "未生成证书，请先运行 gen-certs.sh" >&2; exit 2; }
[ -x "$NGINX_BIN" ] || { echo "找不到 nginx：$NGINX_BIN" >&2; exit 2; }

cat > "$CONF_DIR/nshttps.conf" <<EOF
# ============================================================
# nshttps_gm 统一入口配置（自动生成，勿手改）
# 监听 ${PORT}：国际 TLS + 国密 NTLS 双协议共存（单向认证）
#   /api/*                -> 后端 REST     ${UPSTREAM}
#   音视频播放流(HLS/FLV/TS) -> ZLMediaKit   ${ZLM}
#   /index/api/, /webrtc/  -> ZLMediaKit   ${ZLM}
#   /                     -> 后端前端托管  ${UPSTREAM}
# 说明: 内部服务仅监听 127.0.0.1, 对外统一由本入口承载
# ============================================================
worker_processes auto;
pid ${PID_DIR}/nginx.pid;
error_log ${LOG_DIR}/error.log warn;

events {
    worker_connections 2048;
}

http {
    include       ${MIME_TYPES};
    default_type  application/octet-stream;
    access_log    ${LOG_DIR}/access.log;
    sendfile      on;
    client_body_temp_path ${TMP_DIR}/client_body 2;
    proxy_temp_path       ${TMP_DIR}/proxy 2;
    fastcgi_temp_path     ${TMP_DIR}/fastcgi 2;
    uwsgi_temp_path       ${TMP_DIR}/uwsgi 2;
    scgi_temp_path        ${TMP_DIR}/scgi 2;

    # ---- 国际 TLS：标准 ECC(P-256) 证书 ----
    ssl_certificate      ${CERT_DIR}/server/international.crt;
    ssl_certificate_key  ${CERT_DIR}/server/international.key;

    # ---- 国密 NTLS：SM2 双证书 ----
    ssl_sign_certificate     ${CERT_DIR}/server/server.crt;
    ssl_sign_certificate_key ${CERT_DIR}/server/server.key;
    ssl_enc_certificate      ${CERT_DIR}/server/server_enc.crt;
    ssl_enc_certificate_key  ${CERT_DIR}/server/server_enc.key;
    ssl_trusted_certificate  ${CERT_DIR}/ca/sm2-root.crt;

    ssl_protocols  TLSv1.2 TLSv1.3;
    ssl_ciphers    HIGH:!aNULL:!MD5:!RC4:!kRSA:!kSRP:!kPSK:@STRENGTH;
    ssl_prefer_server_ciphers on;
    enable_ntls on;

    ssl_session_cache   shared:NSG:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets on;

    server {
        listen       0.0.0.0:${PORT} ssl;
        server_name  localhost;

        # 后端 REST（含音视频业务 /api/v1/live, /api/v1/media）
        location /api/ {
            proxy_pass http://${UPSTREAM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # ZLMediaKit API / WebRTC 信令 -> ZLM
        location /index/api/ {
            proxy_pass http://${ZLM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # ZLMediaKit WebRTC 播放页 -> ZLM
        location /webrtc/ {
            proxy_pass http://${ZLM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # ZLMediaKit HLS m3u8 -> ZLM
        location ~ ^/[A-Za-z0-9_]+/[^/]+/hls\.m3u8$ {
            proxy_pass http://${ZLM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # ZLMediaKit HLS ts 分片 -> ZLM
        location ~ ^/[A-Za-z0-9_]+/[^/]+/[0-9-]+/[0-9]+/[^/]+\.ts$ {
            proxy_pass http://${ZLM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # ZLMediaKit HTTP-FLV/TS/FMP4/MP4 -> ZLM
        location ~ ^/[A-Za-z0-9_]+/[^/]+\.(flv|ts|mp4|fmp4)$ {
            proxy_pass http://${ZLM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }

        # 前端托管 / 其他 -> 后端
        location / {
            proxy_pass http://${UPSTREAM};
            proxy_set_header Host              \$host;
            proxy_set_header X-Real-IP         \$remote_addr;
            proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
        }
    }
}
EOF

echo "[nshttps_gm] 已生成配置：$CONF_DIR/nshttps.conf"
echo "[nshttps_gm] 监听端口  : ${PORT}（国际TLS + 国密NTLS 共存）"
echo "[nshttps_gm] 后端REST  : ${UPSTREAM}"
echo "[nshttps_gm] 音视频播放: ${ZLM}"
echo "[nshttps_gm] 国际TLS证书 : ${CERT_DIR}/server/international.crt"
echo "[nshttps_gm] 国密NTLS证书: ${CERT_DIR}/server/server.crt + server_enc.crt"