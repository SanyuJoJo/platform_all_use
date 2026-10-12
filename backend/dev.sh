#!/usr/bin/env bash
#
# 开发调试总入口（默认前台运行）
#
# 用法：
#   ./dev.sh                  # 编译 + 前台启动（Ctrl+C 退出）
#   ./dev.sh fg               # 同默认
#   ./dev.sh bg               # 编译 + 后台启动
#   ./dev.sh build            # 只编译
#   ./dev.sh standalone       # 编译 + 启动 standalone 插件 + 前台 host
#   ./dev.sh plugin echo      # 编译 + 单独启动 echo 插件
#   ./dev.sh stop
#   ./dev.sh status
#   ./dev.sh logs
#   ./dev.sh cross            # 交叉编译全平台到 build/
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

CMD="${1:-run}"
shift || true
REST_ARGS=("$@")

case "$CMD" in
    run|fg)     make run ;;
    bg)         make run-bg ;;
    build|dev)  make all ;;
    standalone) make all && bash bin/scripts/start.sh standalone ;;
    plugin)     make all && bash bin/scripts/start.sh plugin "${REST_ARGS[*]:-}" ;;
    stop)       make stop ;;
    status)     make status ;;
    logs)       make logs ;;
    cross)      make cross ;;
    *)
        cat <<EOF
用法: $0 <命令>

  run / fg            编译 + 前台启动（默认，Ctrl+C 退出）
  bg                  编译 + 后台启动
  build               只编译，不启动
  standalone          编译 + 启动 standalone 插件 + 前台 host
  plugin <name>       编译 + 单独启动某个插件
  stop                停止全部
  status              查看运行状态
  logs                实时查看日志
  cross               交叉编译全平台到 build/
EOF
        exit 1
        ;;
esac
