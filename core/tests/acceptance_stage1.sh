#!/usr/bin/env bash
# core 基础与铜锁编译阶段1 验收测试
# 用法:
#   cd /code/platform_all_use/core
#   bash tests/acceptance_stage1.sh
#
# 可选环境变量:
#   CORE_ROOT=/code/platform_all_use/core
#   OPENSSL_BIN=/opt/core/run/bin/openssl
#   BUILD_OPENSSL_BIN=$CORE_ROOT/libs/bin/tongsuo/bin/openssl
#   KEYCRYPT_BIN=$CORE_ROOT/src/tool/bin/keycrypt
#   TONSUO_MIN_VERSION=8.5.0
#   RUN_CLEAN_BUILD=1   # 执行干净编译，耗时且会 distclean
#   TMPDIR=/tmp

set -u

CORE_ROOT="${CORE_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
OPENSSL_BIN="${OPENSSL_BIN:-/opt/core/run/bin/openssl}"
BUILD_OPENSSL_BIN="${BUILD_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
KEYCRYPT_BIN="${KEYCRYPT_BIN:-$CORE_ROOT/src/tool/bin/keycrypt}"
MIN_VERSION="${TONSUO_MIN_VERSION:-8.5.0}"
RUN_CLEAN_BUILD="${RUN_CLEAN_BUILD:-0}"

REAL_CORE_ROOT="$CORE_ROOT"
TMP="$(mktemp -d "${TMPDIR:-/tmp}/core-stage1-test.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT

PASS=0
FAIL=0
SKIP=0

ok()   { PASS=$((PASS+1)); printf 'PASS  %s\n' "$1"; }
fail() { FAIL=$((FAIL+1)); printf 'FAIL  %s\n' "$1"; }
skip() { SKIP=$((SKIP+1)); printf 'SKIP  %s\n' "$1"; }

if ! command -v python3 >/dev/null 2>&1; then
  echo "需要 python3 来校验 JSON"
  exit 2
fi

need_file() {
  if [ -f "$1" ]; then ok "存在文件: $1"; else fail "缺少文件: $1"; fi
}

need_exec() {
  if [ -x "$1" ]; then ok "可执行: $1"; else fail "不可执行: $1"; fi
}

mode_of() {
  stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1" 2>/dev/null || echo '?'
}

check_mode() {
  path="$1"; want="$2"
  if [ ! -e "$path" ]; then
    fail "权限检查: 缺少 $path"
    return
  fi
  got="$(mode_of "$path")"
  if [ "$got" = "$want" ]; then
    ok "权限 $got $path"
  else
    fail "权限 $path 期望 $want 实际 $got"
  fi
}

json_ok() {
  python3 - "$1" <<'PY'
import json, sys
try:
    json.load(open(sys.argv[1]))
except Exception as e:
    print(e)
    sys.exit(1)
PY
}

json_get() {
  python3 - "$1" "$2" <<'PY'
import json, sys
try:
    d = json.load(open(sys.argv[1]))
except Exception:
    sys.exit(2)
v = d.get(sys.argv[2], "")
print("" if v is None else v)
PY
}

make_fake_openssl() {
  path="$1"; text="$2"
  cat > "$path" <<EOF
#!/bin/sh
echo "$text"
EOF
  chmod +x "$path"
}

run_version_case() {
  desc="$1"; bin="$2"; min="$3"; want_rc="$4"; want_code="$5"
  out="$TMP/vc-out.json"
  err="$TMP/vc-err.txt"
  set +e
  "$CORE_ROOT/sbin/lib/version_check.sh" --openssl "$bin" --min "$min" --json >"$out" 2>"$err"
  rc=$?
  set -e
  if [ "$rc" -ne "$want_rc" ]; then
    fail "$desc: 退出码期望 $want_rc 实际 $rc; stderr=$(cat "$err")"
    return
  fi
  if ! json_ok "$out" >/dev/null 2>&1; then
    fail "$desc: stdout 不是合法 JSON: $(cat "$out")"
    return
  fi
  code="$(json_get "$out" code 2>/dev/null || true)"
  if [ "$code" != "$want_code" ]; then
    fail "$desc: code 期望 $want_code 实际 $code"
    return
  fi
  ok "$desc"
}

echo "== 基础文件 =="
need_file "$CORE_ROOT/Makefile"
need_file "$CORE_ROOT/libs/Makefile"
need_file "$CORE_ROOT/src/tool/Makefile"
need_file "$CORE_ROOT/sbin/dispatch.sh"
need_file "$CORE_ROOT/sbin/lib/common.sh"
need_file "$CORE_ROOT/sbin/lib/version_check.sh"
need_file "$CORE_ROOT/sbin/lib/key_store.sh"
need_file "$CORE_ROOT/sbin/lib/audit.sh"
need_file "$CORE_ROOT/conf/core.conf"
need_file "$CORE_ROOT/tests/test_version_check.sh"
need_file "$CORE_ROOT/tests/test_libs_makefile.sh"
need_exec "$CORE_ROOT/sbin/dispatch.sh"
need_exec "$CORE_ROOT/sbin/lib/version_check.sh"

echo "== 权限 =="
check_mode "$CORE_ROOT" 755
check_mode "$CORE_ROOT/conf" 750
check_mode "$CORE_ROOT/logs" 750
check_mode "$CORE_ROOT/data/keys" 700
check_mode "$CORE_ROOT/tmp" 700
check_mode "$CORE_ROOT/conf/core.conf" 640
check_mode "$CORE_ROOT/libs/bin/tongsuo" 755

echo "== 铜锁源码 =="
if [ -d "$CORE_ROOT/libs/src/tongsuo" ]; then
  ok "源码目录存在: $CORE_ROOT/libs/src/tongsuo"
  if [ -n "$(find "$CORE_ROOT/libs/src/tongsuo" -name .git -print -quit 2>/dev/null)" ]; then
    fail "源码中仍存在 .git"
  else
    ok "源码无 .git"
  fi
  need_file "$CORE_ROOT/libs/src/tongsuo/build.sh"
  need_exec "$CORE_ROOT/libs/src/tongsuo/build.sh"
else
  fail "缺少源码目录: $CORE_ROOT/libs/src/tongsuo"
fi

if [ -f "$CORE_ROOT/libs/src/tongsuo/build.sh" ] && [ -f "$CORE_ROOT/libs/Makefile" ]; then
  if grep -q 'TONSUO_PREFIX' "$CORE_ROOT/libs/src/tongsuo/build.sh" && \
     ! grep -q 'TONSUO_PREFIX=' "$CORE_ROOT/libs/Makefile"; then
    fail "build.sh 需要 TONSUO_PREFIX，但 libs/Makefile 未传 TONSUO_PREFIX（可能传了 TONSUO_BUILD_PREFIX）"
  else
    ok "build.sh 与 libs/Makefile 前缀变量未发现明显不匹配"
  fi
fi

if [ -f "$CORE_ROOT/src/tool/Makefile" ]; then
  if grep -q 'rm -f bin/' "$CORE_ROOT/src/tool/Makefile"; then
    fail "tool Makefile clean 使用 rm -f bin/，删不掉目录"
  else
    ok "tool Makefile clean 未发现 rm -f bin/"
  fi
fi

echo "== 版本校验边界 =="
make_fake_openssl "$TMP/openssl-ok" "Tongsuo 8.5.0"
make_fake_openssl "$TMP/openssl-ok2" "OpenSSL 8.5.1"
make_fake_openssl "$TMP/openssl-low" "Tongsuo 8.4.9"
make_fake_openssl "$TMP/openssl-bad" "not a version"

run_version_case "Tongsuo 8.5.0 应通过" "$TMP/openssl-ok" 8.5.0 0 OK
run_version_case "OpenSSL 8.5.1 应通过" "$TMP/openssl-ok2" 8.5.0 0 OK
run_version_case "Tongsuo 8.4.9 应失败退出12" "$TMP/openssl-low" 8.5.0 12 VERSION_UNSUPPORTED
run_version_case "不存在二进制应失败退出12" "$TMP/no-such-openssl" 8.5.0 12 OPENSSL_BIN_NOT_FOUND
run_version_case "无法解析版本应失败退出12" "$TMP/openssl-bad" 8.5.0 12 VERSION_PARSE_FAILED

echo "== 真实铜锁二进制 =="
for bin in "$BUILD_OPENSSL_BIN" "$OPENSSL_BIN"; do
  if [ -x "$bin" ]; then
    ok "存在可执行: $bin"
    raw="$("$bin" version 2>/dev/null || true)"
    if printf '%s' "$raw" | grep -Eq 'Tongsuo|OpenSSL'; then
      ok "$bin version 输出: $raw"
    else
      fail "$bin version 输出异常: $raw"
    fi

    out="$TMP/real-vc.json"
    set +e
    TONSUO_OPENSSL_BIN="$bin" TONSUO_MIN_VERSION="$MIN_VERSION" \
      "$CORE_ROOT/sbin/lib/version_check.sh" --json >"$out" 2>"$TMP/real-vc.err"
    rc=$?
    set -e
    if [ "$rc" -eq 0 ]; then
      ok "真实版本校验通过: $bin"
    else
      fail "真实版本校验失败: $bin rc=$rc err=$(cat "$TMP/real-vc.err")"
    fi

    if command -v ldd >/dev/null 2>&1; then
      if ldd "$bin" 2>/dev/null | grep -q 'not found'; then
        fail "$bin 动态库缺失: $(ldd "$bin" | grep 'not found')"
      else
        ok "$bin 动态库可解析"
      fi
    fi
  else
    fail "缺少可执行: $bin"
  fi
done

if command -v readelf >/dev/null 2>&1 && [ -x "$OPENSSL_BIN" ]; then
  rpath="$(readelf -d "$OPENSSL_BIN" 2>/dev/null | grep -E 'RPATH|RUNPATH' || true)"
  if printf '%s' "$rpath" | grep -qE '/opt/core/libs/tongsuo|libs/bin/tongsuo/libs'; then
    ok "RPATH/RUNPATH 指向铜锁库"
  else
    fail "RPATH/RUNPATH 未指向铜锁库: $rpath"
  fi
fi

echo "== dispatch 端到端 =="
if [ -x "$CORE_ROOT/sbin/dispatch.sh" ]; then
  echo '{}' > "$TMP/in.json"

  out="$TMP/dispatch-low.json"
  set +e
  TONSUO_OPENSSL_BIN="$TMP/openssl-low" TONSUO_MIN_VERSION=8.5.0 \
    "$CORE_ROOT/sbin/dispatch.sh" --op test --in "$TMP/in.json" --out "$out" \
    >"$TMP/dispatch-low.stdout" 2>"$TMP/dispatch-low.stderr"
  rc=$?
  set -e
  if [ "$rc" -eq 12 ] && json_ok "$out" >/dev/null 2>&1 && \
     [ "$(json_get "$out" code 2>/dev/null)" = "VERSION_UNSUPPORTED" ]; then
    ok "dispatch 版本失败路径 exit12/VERSION_UNSUPPORTED"
  else
    fail "dispatch 版本失败路径异常 rc=$rc out=$(cat "$out" 2>/dev/null)"
  fi

  out2="$TMP/dispatch-ok.json"
  set +e
  TONSUO_OPENSSL_BIN="$TMP/openssl-ok" TONSUO_MIN_VERSION=8.5.0 \
    "$CORE_ROOT/sbin/dispatch.sh" --op test --in "$TMP/in.json" --out "$out2" \
    >"$TMP/dispatch-ok.stdout" 2>"$TMP/dispatch-ok.stderr"
  rc2=$?
  set -e
  if [ "$rc2" -eq 99 ] && json_ok "$out2" >/dev/null 2>&1 && \
     [ "$(json_get "$out2" code 2>/dev/null)" = "INTERNAL_ERROR" ]; then
    ok "dispatch 未实现操作占位 exit99/INTERNAL_ERROR"
  else
    fail "dispatch 未实现操作占位异常 rc=$rc2 out=$(cat "$out2" 2>/dev/null)"
  fi
else
  fail "dispatch.sh 不可执行"
fi

echo "== 密钥加密存储（keycrypt 回环） =="
if [ ! -x "$KEYCRYPT_BIN" ]; then
  fail "keycrypt 未编译: $KEYCRYPT_BIN（先执行 make -C src/tool all）"
else
  ok "keycrypt 可执行: $KEYCRYPT_BIN"

  master="$TMP/master.key"
  if [ -x "$BUILD_OPENSSL_BIN" ]; then
    "$BUILD_OPENSSL_BIN" rand -hex 32 > "$master"
  else
    head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' > "$master"
  fi
  chmod 0600 "$master"

  printf 'test-secret-payload\n' > "$TMP/kplain.bin"

  set +e
  "$KEYCRYPT_BIN" encrypt --master "$master" --in "$TMP/kplain.bin" --out "$TMP/kplain.enc" \
    >"$TMP/kc-enc.out" 2>"$TMP/kc-enc.err"
  rc_enc=$?
  set -e
  if [ "$rc_enc" -eq 0 ] && [ -s "$TMP/kplain.enc" ]; then
    ok "keycrypt encrypt 成功"
  else
    fail "keycrypt encrypt 失败 rc=$rc_enc err=$(cat "$TMP/kc-enc.err")"
  fi

  set +e
  "$KEYCRYPT_BIN" decrypt --master "$master" --in "$TMP/kplain.enc" --out "$TMP/kplain.dec" \
    >"$TMP/kc-dec.out" 2>"$TMP/kc-dec.err"
  rc_dec=$?
  set -e
  if [ "$rc_dec" -eq 0 ] && cmp -s "$TMP/kplain.bin" "$TMP/kplain.dec"; then
    ok "keycrypt 加解密回环一致"
  else
    fail "keycrypt 解密回环失败 rc=$rc_dec err=$(cat "$TMP/kc-dec.err")"
  fi

  # 篡改密文，期望解密失败
  cp "$TMP/kplain.enc" "$TMP/kplain.tamper"
  sz="$(wc -c < "$TMP/kplain.tamper")"
  if [ "$sz" -gt 20 ]; then
    pos=$((sz - 20))
    printf '\xff' | dd of="$TMP/kplain.tamper" bs=1 seek="$pos" count=1 conv=notrunc 2>/dev/null
    set +e
    "$KEYCRYPT_BIN" decrypt --master "$master" --in "$TMP/kplain.tamper" --out "$TMP/kplain.bad" \
      >"$TMP/kc-tamper.out" 2>"$TMP/kc-tamper.err"
    rc_tamper=$?
    set -e
    if [ "$rc_tamper" -ne 0 ]; then
      ok "keycrypt 篡改密文被拒绝（GCM 认证失败）"
    else
      fail "keycrypt 篡改密文未被拒绝"
    fi
  else
    skip "密文太短，跳过篡改测试"
  fi

  # 错误主密钥，期望解密失败
  other="$TMP/other.key"
  if [ -x "$BUILD_OPENSSL_BIN" ]; then
    "$BUILD_OPENSSL_BIN" rand -hex 32 > "$other"
  else
    head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' > "$other"
  fi
  chmod 0600 "$other"
  set +e
  "$KEYCRYPT_BIN" decrypt --master "$other" --in "$TMP/kplain.enc" --out "$TMP/kplain.wrong" \
    >"$TMP/kc-wrong.out" 2>"$TMP/kc-wrong.err"
  rc_wrong=$?
  set -e
  if [ "$rc_wrong" -ne 0 ]; then
    ok "keycrypt 错误主密钥被拒绝"
  else
    fail "keycrypt 错误主密钥未被拒绝"
  fi

  # 通过 key_store.sh 再走一遍
  if [ -f "$REAL_CORE_ROOT/sbin/lib/key_store.sh" ]; then
    mkdir -p "$TMP/ks-root/data/keys" "$TMP/ks-root/logs"
    cp "$master" "$TMP/ks-root/data/keys/master.key"
    chmod 0700 "$TMP/ks-root/data/keys"
    chmod 0600 "$TMP/ks-root/data/keys/master.key"

    if (
      CORE_ROOT="$TMP/ks-root"
      MASTER_KEY_PATH="$TMP/ks-root/data/keys/master.key"
      KEYCRYPT_BIN="$KEYCRYPT_BIN"
      TONSUO_OPENSSL_BIN="$BUILD_OPENSSL_BIN"
      . "$REAL_CORE_ROOT/sbin/lib/key_store.sh"
      key_store_encrypt "$TMP/kplain.bin" "$TMP/ks.enc" &&
      key_store_decrypt "$TMP/ks.enc" "$TMP/ks.dec"
    ) >"$TMP/ks.out" 2>"$TMP/ks.err"; then
      if cmp -s "$TMP/kplain.bin" "$TMP/ks.dec"; then
        ok "key_store.sh 加解密回环一致"
      else
        fail "key_store.sh 解密结果不一致"
      fi
      if [ "$(mode_of "$TMP/ks-root/data/keys/master.key")" = "600" ]; then
        ok "key_store.sh master.key 权限 600"
      else
        fail "key_store.sh master.key 权限 $(mode_of "$TMP/ks-root/data/keys/master.key")"
      fi
      if [ "$(mode_of "$TMP/ks.enc")" = "600" ]; then
        ok "key_store.sh 密文文件权限 600"
      else
        fail "key_store.sh 密文文件权限 $(mode_of "$TMP/ks.enc")"
      fi
    else
      fail "key_store.sh 调用失败: $(cat "$TMP/ks.err")"
    fi
  else
    fail "缺少 key_store.sh: $REAL_CORE_ROOT/sbin/lib/key_store.sh"
  fi
fi

if [ -f "$CORE_ROOT/data/keys/master.key" ]; then
  check_mode "$CORE_ROOT/data/keys/master.key" 600
else
  skip "主密钥未生成，跳过 master.key 权限检查"
fi

echo "== 审计日志 =="
if [ -f "$CORE_ROOT/sbin/lib/audit.sh" ]; then
  mkdir -p "$TMP/audit-root/logs"
  if (
    CORE_ROOT="$TMP/audit-root"
    . "$REAL_CORE_ROOT/sbin/lib/audit.sh"
    audit_write "req-1" "ca.create" "SUCCESS" 12 "" "task-1" "SM2" "sha256:abc" "v1"
  ) >/dev/null 2>"$TMP/audit.err"; then
    log="$TMP/audit-root/logs/audit.jsonl"
    if [ -f "$log" ] && json_ok "$log" >/dev/null 2>&1; then
      ok "audit_write 输出合法 JSON"
    else
      fail "audit_write 输出不是合法 JSON: $(cat "$log" 2>/dev/null)"
    fi
  else
    fail "audit_write 调用失败: $(cat "$TMP/audit.err")"
  fi

  mkdir -p "$TMP/audit-escape/logs"
  if (
    CORE_ROOT="$TMP/audit-escape"
    . "$REAL_CORE_ROOT/sbin/lib/audit.sh"
    audit_write 'req-1' 'ca.create' 'SUCCESS' 12 '' 'task"quote' 'SM2' 'sha256:abc' 'v1'
  ) >/dev/null 2>"$TMP/audit-escape.err"; then
    log="$TMP/audit-escape/logs/audit.jsonl"
    if [ -f "$log" ] && json_ok "$log" >/dev/null 2>&1; then
      ok "audit_write 字段含引号仍输出合法 JSON"
    else
      fail "audit_write 字段含引号破坏 JSON: $(cat "$log" 2>/dev/null)"
    fi
  else
    fail "audit_write 转义测试调用失败: $(cat "$TMP/audit-escape.err")"
  fi
else
  fail "缺少 audit.sh"
fi

echo "== 配置读取 =="
if [ -f "$CORE_ROOT/sbin/lib/common.sh" ]; then
  if (
    CORE_ROOT="$CORE_ROOT"
    CORE_CONF="$CORE_ROOT/conf/core.conf"
    . "$CORE_ROOT/sbin/lib/common.sh"
    core_conf_get "tongsuo_min_version" "8.5.0"
  ) >"$TMP/conf.out" 2>"$TMP/conf.err"; then
    val="$(cat "$TMP/conf.out")"
    if [ "$val" = "8.5.0" ]; then
      ok "core_conf_get 读取 tongsuo_min_version"
    else
      fail "core_conf_get 返回异常: $val"
    fi
  else
    fail "core_conf_get 调用失败: $(cat "$TMP/conf.err")"
  fi
fi

echo "== 安全禁止项 =="
hits="$(grep -RIn --exclude-dir=.git -E '\bshell=True\b|\beval\b|\bsh -c\b' \
        "$CORE_ROOT/sbin" "$CORE_ROOT/src" 2>/dev/null || true)"
if [ -n "$hits" ]; then
  fail "发现 shell=True/eval/sh -c 等可疑用法："
  printf '%s\n' "$hits"
else
  ok "未发现 shell=True/eval/sh -c"
fi

echo "== 项目自带测试 =="
for t in "$CORE_ROOT/tests/test_version_check.sh" \
         "$CORE_ROOT/tests/test_libs_makefile.sh" \
         "$CORE_ROOT/tests/test_keycrypt.sh"; do
  if [ -x "$t" ]; then
    name="$(basename "$t")"
    if "$t" >"$TMP/$name.log" 2>&1; then
      ok "自带测试通过: $name"
    else
      fail "自带测试失败: $name 日志: $(tail -n 20 "$TMP/$name.log")"
    fi
  else
    if [ -f "$t" ]; then
      fail "自带测试不可执行: $t"
    else
      skip "自带测试不存在: $t"
    fi
  fi
done

echo "== 干净构建（可选） =="
if [ "$RUN_CLEAN_BUILD" = "1" ]; then
  TMP_DEPLOY="$TMP/deploy"
  if make -C "$CORE_ROOT/libs" distclean >"$TMP/build-distclean.log" 2>&1; then
    ok "libs distclean 完成"
  else
    fail "libs distclean 失败"
  fi

  if make -C "$CORE_ROOT/libs" all DEPLOY_ROOT="$TMP_DEPLOY" \
      JOBS="$(nproc 2>/dev/null || echo 2)" >"$TMP/build.log" 2>&1; then
    ok "libs 干净构建完成"
    if [ -x "$TMP_DEPLOY/run/bin/openssl" ]; then
      ok "部署 openssl 存在"
    else
      fail "部署 openssl 不存在"
    fi
  else
    fail "libs 干净构建失败，日志: $(tail -n 20 "$TMP/build.log")"
  fi
else
  skip "RUN_CLEAN_BUILD=1 时执行干净编译"
fi

echo
echo "SUMMARY: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
