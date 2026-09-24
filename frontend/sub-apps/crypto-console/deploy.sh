#!/usr/bin/env bash
set -e

APP_NAME="crypto-console"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(cd "$SRC_DIR/../../deploy/sub-apps" && pwd)/$APP_NAME"

echo "==> 构建 $APP_NAME"
cd "$SRC_DIR"
pnpm build:prod

echo "==> 复制到 $DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR"
rm -rf "$DEPLOY_DIR"/*
cp -r dist/* "$DEPLOY_DIR/"

echo "==> 完成"
ls -la "$DEPLOY_DIR"
