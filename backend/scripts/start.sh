#!/usr/bin/env bash
#
# 开发调试一键启动脚本
#
# 用法：
#   ./start.sh                # 前台启动 host（默认）
#   ./start.sh fg             # 同上
#   ./start.sh start          # 后台启动 host
#   ./start.sh standalone     # 启动 standalone 插件后，前台启动 host
#   ./start.sh plugin <name>  # 单独启动某个插件
#   ./start.sh stop
#   ./start.sh status
#
# 配置：
#   默认使用 ../config.json（即 bin/config.json）。
#   可通过环境变量 HOST_CONFIG 指定其他路径：
#       HOST_CONFIG=/etc/backend/config.json ./start.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PLUGIN_DIR="$BIN_DIR/plugins"
DATA_DIR="$BIN_DIR/data"
LOG_DIR="$DATA_DIR/logs"
PID_DIR="$DATA_DIR/pids"

# 默认配置文件 = bin/config.json
CONFIG_FILE="${HOST_CONFIG:-$BIN_DIR/config.json}"

UNAME_S="$(uname -s || echo unknown)"
case "$UNAME_S" in
    MINGW*|MSYS*|CYGWIN*) EXT=".exe";;
    *)                    EXT="";;
esac

HOST_BIN="$BIN_DIR/host$EXT"

mkdir -p "$LOG_DIR" "$PID_DIR" "$DATA_DIR/backups"

# —— 工具函数 ——

read_manifest_field() {
    local file="$1" key="$2"
    sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" "$file" | head -n1
}

pid_alive() {
    local pid="$1"
    [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null
}

# —— 从配置文件读取 host/port，用于日志提示 ——
peek_config() {
    local file="$1"
    if [ ! -f "$file" ]; then
        echo "(配置文件不存在: $file)"
        return
    fi
    local host port
    host="$(sed -n 's/.*"host"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$file" | head -n1)"
    port="$(sed -n 's/.*"port"[[:space:]]*:[[:space:]]*\([0-9]\+\).*/\1/p' "$file" | head -n1)"
    echo "host=${host:-0.0.0.0} port=${port:-8080}"
}

# —— 无条件清理 ——
force_cleanup() {
    if [ -d "$PID_DIR" ] && ls "$PID_DIR"/*.pid >/dev/null 2>&1; then
        stop_all 2>/dev/null || true
    fi
    pkill -f "$BIN_DIR/host" 2>/dev/null || true
    pkill -f "$BIN_DIR/plugins/" 2>/dev/null || true
    pkill -f "plugin-python-hello/main.py" 2>/dev/null || true

    if command -v lsof >/dev/null 2>&1; then
        for port in 19100 9100; do
            local pids
            pids="$(lsof -ti :$port 2>/dev/null || true)"
            [ -n "$pids" ] && kill -9 $pids 2>/dev/null || true
        done
    fi
    sleep 0.5
}

# —— 前台启动 ——
start_host_fg() {
    if [ ! -f "$HOST_BIN" ]; then
        echo "[start] 未找到主程序: $HOST_BIN"
        echo "        请先运行: make all"
        exit 1
    fi
    if [ -f "$PID_DIR/host.pid" ] && pid_alive "$(cat "$PID_DIR/host.pid")"; then
        echo "[start] 已有后台 host 在运行 (PID: $(cat "$PID_DIR/host.pid"))"
        echo "        如要前台启动，请先: ./start.sh stop"
        exit 1
    fi

    echo "[start] 前台启动 host（Ctrl+C 退出）"
    echo "[start] 配置文件: $CONFIG_FILE"
    echo "[start] 配置内容: $(peek_config "$CONFIG_FILE")"
    echo "[start] 日志: $LOG_DIR/host.log"
    echo "-------------------------------------------------------------"

    HOST_CONFIG="$CONFIG_FILE" exec "$HOST_BIN" \
        > >(tee -a "$LOG_DIR/host.log") 2>&1
}

# —— 后台启动 ——
start_host_bg() {
    if [ -f "$PID_DIR/host.pid" ] && pid_alive "$(cat "$PID_DIR/host.pid")"; then
        echo "[start] host 已在运行 (PID: $(cat "$PID_DIR/host.pid"))"
        return 0
    fi
    [ -f "$HOST_BIN" ] || { echo "[start] 未找到主程序: $HOST_BIN"; exit 1; }

    echo "[start] 后台启动 host"
    echo "[start] 配置文件: $CONFIG_FILE"
    echo "[start] 配置内容: $(peek_config "$CONFIG_FILE")"
    HOST_CONFIG="$CONFIG_FILE" nohup "$HOST_BIN" \
        > "$LOG_DIR/host.log" 2>&1 &
    echo $! > "$PID_DIR/host.pid"
    echo "[start] host 已启动 (PID: $!)  日志: $LOG_DIR/host.log"
}

# —— 单独启动插件 ——
start_plugin_standalone() {
    local name="$1"
    local dir="$PLUGIN_DIR/$name"
    [ -d "$dir" ] || { echo "[start] 插件目录不存在: $dir"; exit 1; }

    local manifest="$dir/manifest.json"
    [ -f "$manifest" ] || { echo "[start] 缺少 manifest: $manifest"; exit 1; }

    local entry runtime port
    entry="$(read_manifest_field "$manifest" entrypoint)"
    runtime="$(read_manifest_field "$manifest" runtime)"
    port="$(sed -n 's/.*"port"[[:space:]]*:[[:space:]]*\([0-9]\+\).*/\1/p' "$manifest" | head -n1)"

    if [ -f "$PID_DIR/$name.pid" ] && pid_alive "$(cat "$PID_DIR/$name.pid")"; then
        echo "[start] 插件 $name 已在运行 (PID: $(cat "$PID_DIR/$name.pid"))"
        return 0
    fi

    echo "[start] 启动插件 $name (runtime=$runtime, entry=$entry, port=$port)"

    case "$runtime" in
        http)
            if [[ "$entry" == *.py ]] || [[ "$entry" == python* ]]; then
                PLUGIN_STANDALONE=true PLUGIN_PORT="$port" \
                    nohup python3 "$dir/main.py" \
                    > "$LOG_DIR/$name.log" 2>&1 &
            else
                PLUGIN_STANDALONE=true PLUGIN_PORT="$port" \
                    nohup "$dir/$entry$EXT" \
                    > "$LOG_DIR/$name.log" 2>&1 &
            fi
            ;;
        go-plugin)
            PLUGIN_STANDALONE=true \
                nohup "$dir/$entry$EXT" \
                > "$LOG_DIR/$name.log" 2>&1 &
            ;;
        *)
            echo "[start] 未知 runtime: $runtime"; exit 1;;
    esac

    echo $! > "$PID_DIR/$name.pid"
    echo "[start] 插件 $name 已启动 (PID: $!)  日志: $LOG_DIR/$name.log"
}

# —— 停止 ——
stop_one() {
    local name="$1"
    local pid_file="$PID_DIR/$name.pid"
    [ -f "$pid_file" ] || return 0
    local pid
    pid="$(cat "$pid_file")"
    if pid_alive "$pid"; then
        echo "[stop] 停止 $name (PID: $pid)"
        kill "$pid" 2>/dev/null || true
        for _ in $(seq 1 20); do
            pid_alive "$pid" || break
            sleep 0.2
        done
        if pid_alive "$pid"; then
            echo "[stop] 强制终止 $name"
            kill -9 "$pid" 2>/dev/null || true
        fi
    fi
    rm -f "$pid_file"
}

stop_all() {
    stop_one host
    for d in "$PLUGIN_DIR"/*/; do
        [ -d "$d" ] || continue
        stop_one "$(basename "$d")"
    done
    echo "[stop] 全部已停止"
}

# —— 状态 ——
status_all() {
    local any=0
    for pid_file in "$PID_DIR"/*.pid; do
        [ -f "$pid_file" ] || continue
        any=1
        local name pid
        name="$(basename "$pid_file" .pid)"
        pid="$(cat "$pid_file")"
        if pid_alive "$pid"; then
            echo "[运行中] $name  PID=$pid"
        else
            echo "[已停止] $name  (stale pid file)"
        fi
    done
    [ "$any" = "1" ] || echo "[status] 无后台进程在运行"

    if pgrep -f "$BIN_DIR/host" >/dev/null 2>&1; then
        echo "[前台] 检测到 host 正在运行:"
        pgrep -af "$BIN_DIR/host" || true
    fi

    echo "[status] 配置文件: $CONFIG_FILE"
    echo "[status] 配置内容: $(peek_config "$CONFIG_FILE")"
}

# —— 入口 ——
cmd="${1:-fg}"
shift || true

case "$cmd" in
    fg|"")        force_cleanup; start_host_fg ;;
    start|bg)     force_cleanup; start_host_bg ;;
    standalone)
        force_cleanup
        for d in "$PLUGIN_DIR"/*/; do
            [ -d "$d" ] || continue
            name="$(basename "$d")"
            manifest="$d/manifest.json"
            [ -f "$manifest" ] || continue
            mode="$(read_manifest_field "$manifest" mode)"
            if [ "$mode" = "standalone" ]; then
                start_plugin_standalone "$name"
            fi
        done
        sleep 1.5
        start_host_fg
        ;;
    plugin)
        [ -n "${1:-}" ] || { echo "用法: $0 plugin <name>"; exit 1; }
        start_plugin_standalone "$1"
        ;;
    stop)    stop_all ;;
    status)  status_all ;;
    *)
        echo "用法: $0 [fg|start|standalone|plugin <name>|stop|status]"
        echo "环境变量: HOST_CONFIG=<配置文件路径>"
        exit 1
        ;;
esac
