#!/bin/sh
# core/tests/test_p1p2_ops.sh
# P1-P2 密码操作交付级回归测试
#
# 覆盖:
#   cert.parse          - 正常/损坏/路径穿越/不存在
#   cert.convert        - PEM→DER / PEM→PKCS12 / 口令权限校验
#   key.manage          - generate / export(默认拒绝) / export(放行) / delete / 非法 action
#   crypto.service      - sign / 路径穿越 / key 不存在 / 非法 action
#   chain.verify        - 自签通过 / 篡改 / 不存在
#   batch.execute       - 正常 / 部分失败 / 递归防护 / 空列表
#   ssl.config.generate - nginx / 非法 server_type
#   pqc.cert.create     - ML-DSA 软断言 / ML-KEM 用途拒绝 / 非法算法
#   审计脱敏             - params_digest / 各操作审计写入 / 无私钥/口令泄露
#   安全基线             - 无 shell=True/eval/sh -c

set -u

CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DISPATCH="$CORE_ROOT/sbin/dispatch.sh"
OPENSSL_BIN="${TONSUO_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
AUDIT_LOG="$CORE_ROOT/logs/audit.jsonl"

TMP=$(mktemp -d "${TMPDIR:-/tmp}/p1p2-ops.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

PASS=0; FAIL=0; SKIP=0
ok()   { PASS=$((PASS+1)); printf 'PASS  %s\n' "$1"; }
fail() { FAIL=$((FAIL+1)); printf 'FAIL  %s\n' "$1"; }
skip() { SKIP=$((SKIP+1)); printf 'SKIP  %s\n' "$1"; }

command -v jq >/dev/null 2>&1 || { echo "jq required"; exit 2; }
[ -x "$DISPATCH" ]    || { echo "dispatch not executable: $DISPATCH"; exit 2; }
[ -x "$OPENSSL_BIN" ] || { echo "openssl not executable: $OPENSSL_BIN"; exit 2; }

mkdir -p "$CORE_ROOT/tmp"
chmod 0700 "$CORE_ROOT/tmp" 2>/dev/null || true

mode_of() {
  stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1" 2>/dev/null || echo '?'
}

# 记录本轮测试开始前的审计行数，只对新写入的条目做断言
AUDIT_START_LINE=0
if [ -f "$AUDIT_LOG" ]; then
  AUDIT_START_LINE=$(wc -l < "$AUDIT_LOG" | tr -d ' ')
fi
audit_tail() {
  [ -f "$AUDIT_LOG" ] || return 0
  tail -n +"$((AUDIT_START_LINE + 1))" "$AUDIT_LOG"
}

# 统一操作调用与断言
run_op() {
  case_name="$1"; op="$2"; in_file="$3"; expect_rc="$4"; expect_code="$5"
  out="$TMP/$case_name.out.json"
  serr="$TMP/$case_name.stderr"

  set +e
  "$DISPATCH" --op "$op" --in "$in_file" --out "$out" >/dev/null 2>"$serr"
  rc=$?
  set -e

  if [ "$rc" -ne "$expect_rc" ]; then
    fail "$case_name: exit=$rc expected=$expect_rc; stderr=$(tail -n3 "$serr" 2>/dev/null | tr '\n' ' ')"
    return 1
  fi
  if [ ! -f "$out" ]; then
    fail "$case_name: --out missing"; return 1
  fi
  if ! jq -e . "$out" >/dev/null 2>&1; then
    fail "$case_name: --out not valid JSON"; return 1
  fi
  code="$(jq -r '.code // empty' "$out")"
  if [ "$code" != "$expect_code" ]; then
    fail "$case_name: code=$code expected=$expect_code"; return 1
  fi
  ok "$case_name: exit=$rc code=$code"
  return 0
}

# 断言非零退出码（用于路径穿越等）
run_op_nonzero() {
  case_name="$1"; op="$2"; in_file="$3"; expect_rc="$4"
  out="$TMP/$case_name.out.json"
  serr="$TMP/$case_name.stderr"

  set +e
  "$DISPATCH" --op "$op" --in "$in_file" --out "$out" >/dev/null 2>"$serr"
  rc=$?
  set -e

  if [ "$rc" -ne "$expect_rc" ]; then
    fail "$case_name: exit=$rc expected=$expect_rc"
    return 1
  fi
  ok "$case_name: exit=$rc (as expected)"
  return 0
}

echo "============================================"
echo " P1-P2 密码操作回归测试"
echo " CORE_ROOT=$CORE_ROOT"
echo "============================================"

# =============================================================
# 0. 准备测试数据
# =============================================================
echo
echo "---- 准备测试数据 ----"

CA_REL=""
CA_ABS=""
CA_EXISTING="$(ls -t "$CORE_ROOT/data/ca/"*.pem 2>/dev/null | grep -v -- '-chain.pem' | head -n1 || true)"
if [ -n "$CA_EXISTING" ] && [ -f "$CA_EXISTING" ]; then
  CA_ABS="$CA_EXISTING"
  CA_REL="${CA_EXISTING#$CORE_ROOT/}"
  ok "复用已有 CA: $CA_REL"
else
  SETUP_IN="$TMP/setup-ca.json"
  cat > "$SETUP_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-p1p2-setup-ca","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"SM2","subject":{"CN":"P1P2 Test CA","O":"Test"},"validity_days":3650}}
EOF
  if run_op "setup.ca.create" "ca.create" "$SETUP_IN" 0 "OK"; then
    CA_REL="$(jq -r '.data.cert_path' "$TMP/setup.ca.create.out.json")"
    CA_ABS="$CORE_ROOT/$CA_REL"
    ok "已创建测试 CA: $CA_REL"
  else
    fail "无法创建测试 CA，多数 P1-P2 用例将 SKIP"
  fi
fi

PAYLOAD_REL="tmp/test-p1p2-payload.txt"
PASS_REL="tmp/test-p1p2-pass.txt"
printf 'test-payload-for-p1p2\n' > "$CORE_ROOT/$PAYLOAD_REL"
printf 'test-password-12345\n'   > "$CORE_ROOT/$PASS_REL"
chmod 0600 "$CORE_ROOT/$PAYLOAD_REL" "$CORE_ROOT/$PASS_REL"
ok "已准备 payload 与 password_file"

# =============================================================
# 1. cert.parse
# =============================================================
echo
echo "---- cert.parse ----"
if [ -n "$CA_REL" ]; then
  PARSE_IN="$TMP/parse.json"
  cat > "$PARSE_IN" <<EOF
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"req-p1p2-parse","actor":{"type":"platform-backend","id":"t"},"params":{"cert_path":"$CA_REL"}}
EOF
  run_op "cert.parse.ok" "cert.parse" "$PARSE_IN" 0 "OK" || true
  PARSE_OUT="$TMP/cert.parse.ok.out.json"
  if [ -f "$PARSE_OUT" ]; then
    SUBJ="$(jq -r '.data.subject // empty' "$PARSE_OUT")"
    SERIAL="$(jq -r '.data.serial // empty' "$PARSE_OUT")"
    FP="$(jq -r '.data.fingerprint_sha256 // empty' "$PARSE_OUT")"
    NB="$(jq -r '.data.not_before // empty' "$PARSE_OUT")"
    [ -n "$SUBJ" ]   && ok "cert.parse data.subject 非空"     || fail "cert.parse data.subject 为空"
    [ -n "$SERIAL" ] && ok "cert.parse data.serial 非空"      || fail "cert.parse data.serial 为空"
    [ -n "$FP" ]     && ok "cert.parse data.fingerprint 非空" || fail "cert.parse data.fingerprint 为空"
    [ -n "$NB" ]     && ok "cert.parse data.not_before 非空"  || fail "cert.parse data.not_before 为空"
  fi

  # 损坏证书
  printf 'not-a-valid-cert\n' > "$CORE_ROOT/tmp/test-p1p2-bad.pem"
  PARSE_BAD="$TMP/parse-bad.json"
  cat > "$PARSE_BAD" <<'EOF'
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"req-p1p2-parse-bad","actor":{"type":"platform-backend","id":"t"},"params":{"cert_path":"tmp/test-p1p2-bad.pem"}}
EOF
  run_op "cert.parse.bad-content" "cert.parse" "$PARSE_BAD" 8 "CERT_PARSE_FAILED" || true

  # 路径穿越
  PARSE_TRAV="$TMP/parse-trav.json"
  cat > "$PARSE_TRAV" <<'EOF'
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"req-p1p2-parse-trav","actor":{"type":"platform-backend","id":"t"},"params":{"cert_path":"../../../etc/passwd"}}
EOF
  run_op_nonzero "cert.parse.path-traversal" "cert.parse" "$PARSE_TRAV" 4 || true

  # 不存在的证书
  PARSE_NF="$TMP/parse-nf.json"
  cat > "$PARSE_NF" <<'EOF'
{"schema_version":"1.0","operation_id":"cert.parse","request_id":"req-p1p2-parse-nf","actor":{"type":"platform-backend","id":"t"},"params":{"cert_path":"data/ca/does-not-exist.pem"}}
EOF
  run_op "cert.parse.not-found" "cert.parse" "$PARSE_NF" 7 "CERT_NOT_FOUND" || true
else
  skip "cert.parse：无 CA 证书"
fi

# =============================================================
# 2. cert.convert
# =============================================================
echo
echo "---- cert.convert ----"
if [ -n "$CA_REL" ]; then
  # PEM → DER
  CV_DER="$TMP/convert-der.json"
  cat > "$CV_DER" <<EOF
{"schema_version":"1.0","operation_id":"cert.convert","request_id":"req-p1p2-cv-der","actor":{"type":"platform-backend","id":"t"},"params":{"source_format":"PEM","target_format":"DER","source_path":"$CA_REL"}}
EOF
  run_op "cert.convert.PEM2DER" "cert.convert" "$CV_DER" 0 "OK" || true
  CV_DER_OUT="$TMP/cert.convert.PEM2DER.out.json"
  if [ -f "$CV_DER_OUT" ]; then
    DER_REL="$(jq -r '.data.converted_path // empty' "$CV_DER_OUT")"
    if [ -n "$DER_REL" ] && [ -f "$CORE_ROOT/$DER_REL" ]; then
      if "$OPENSSL_BIN" x509 -in "$CORE_ROOT/$DER_REL" -inform DER -noout 2>/dev/null; then
        ok "cert.convert PEM→DER 输出可解析"
      else
        fail "cert.convert PEM→DER 输出无法解析"
      fi
    fi
  fi

  # PEM → PKCS12
  CV_P12="$TMP/convert-p12.json"
  cat > "$CV_P12" <<EOF
{"schema_version":"1.0","operation_id":"cert.convert","request_id":"req-p1p2-cv-p12","actor":{"type":"platform-backend","id":"t"},"params":{"source_format":"PEM","target_format":"PKCS12","source_path":"$CA_REL","password_file":"$PASS_REL"}}
EOF
  run_op "cert.convert.PEM2PKCS12" "cert.convert" "$CV_P12" 0 "OK" || true
  CV_P12_OUT="$TMP/cert.convert.PEM2PKCS12.out.json"
  if [ -f "$CV_P12_OUT" ]; then
    P12_REL="$(jq -r '.data.converted_path // empty' "$CV_P12_OUT")"
    if [ -n "$P12_REL" ] && [ -f "$CORE_ROOT/$P12_REL" ]; then
      if "$OPENSSL_BIN" pkcs12 -in "$CORE_ROOT/$P12_REL" -nokeys -passin "file:$CORE_ROOT/$PASS_REL" -noout 2>/dev/null; then
        ok "cert.convert PKCS12 可用口令打开"
      else
        fail "cert.convert PKCS12 无法用口令打开"
      fi
      m="$(mode_of "$CORE_ROOT/$P12_REL")"
      [ "$m" = "600" ] && ok "cert.convert 输出权限 600" || fail "cert.convert 输出权限 $m"
    fi
  fi

  # password_file 权限非 0600 应被拒绝
  printf 'pw\n' > "$CORE_ROOT/tmp/test-p1p2-badpass.txt"
  chmod 0644 "$CORE_ROOT/tmp/test-p1p2-badpass.txt"
  CV_BP="$TMP/convert-badpass.json"
  cat > "$CV_BP" <<EOF
{"schema_version":"1.0","operation_id":"cert.convert","request_id":"req-p1p2-cv-badpass","actor":{"type":"platform-backend","id":"t"},"params":{"source_format":"PEM","target_format":"PKCS12","source_path":"$CA_REL","password_file":"tmp/test-p1p2-badpass.txt"}}
EOF
  run_op "cert.convert.bad-passfile-mode" "cert.convert" "$CV_BP" 2 "INVALID_PARAM" || true
else
  skip "cert.convert：无 CA 证书"
fi

# =============================================================
# 3. key.manage
# =============================================================
echo
echo "---- key.manage ----"

KM_GEN="$TMP/key-gen.json"
cat > "$KM_GEN" <<'EOF'
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-gen","actor":{"type":"platform-backend","id":"t"},"params":{"action":"generate","algorithm":"SM2"}}
EOF
run_op "key.manage.generate" "key.manage" "$KM_GEN" 0 "OK" || true
KM_GEN_OUT="$TMP/key.manage.generate.out.json"
GEN_KEY_REF="$(jq -r '.data.key_ref // empty' "$KM_GEN_OUT" 2>/dev/null)"
GEN_KEY_ENC="$(jq -r '.data.encrypted_path // empty' "$KM_GEN_OUT" 2>/dev/null)"

if [ -n "$GEN_KEY_REF" ]; then
  ok "key.manage.generate key_ref=$GEN_KEY_REF"
  if [ -f "$GEN_KEY_ENC" ]; then
    ok "key.manage.generate 加密私钥文件存在"
    m="$(mode_of "$GEN_KEY_ENC")"
    [ "$m" = "600" ] && ok "key.manage 加密私钥权限 600" || fail "key.manage 加密私钥权限 $m"
  else
    fail "key.manage 加密私钥文件缺失: $GEN_KEY_ENC"
  fi

  KM_EXP_DENY="$TMP/key-exp-deny.json"
  cat > "$KM_EXP_DENY" <<EOF
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-exp-deny","actor":{"type":"platform-backend","id":"t"},"params":{"action":"export","key_ref":"$GEN_KEY_REF"}}
EOF
  run_op "key.manage.export.deny-default" "key.manage" "$KM_EXP_DENY" 5 "PERMISSION_DENIED" || true

  KM_EXP_ALLOW="$TMP/key-exp-allow.json"
  cat > "$KM_EXP_ALLOW" <<EOF
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-exp-allow","actor":{"type":"platform-backend","id":"t"},"params":{"action":"export","key_ref":"$GEN_KEY_REF","allow_plain_export":true,"export_path":"data/export/${GEN_KEY_REF}.pem"}}
EOF
  run_op "key.manage.export.allow" "key.manage" "$KM_EXP_ALLOW" 0 "OK" || true
  KM_EXP_OUT="$TMP/key.manage.export.allow.out.json"
  if [ -f "$KM_EXP_OUT" ]; then
    EXP_PATH="$(jq -r '.data.export_path // empty' "$KM_EXP_OUT")"
    if [ -n "$EXP_PATH" ] && [ -f "$CORE_ROOT/$EXP_PATH" ]; then
      m="$(mode_of "$CORE_ROOT/$EXP_PATH")"
      [ "$m" = "600" ] && ok "key.manage 导出文件权限 600" || fail "key.manage 导出文件权限 $m"
    fi
  fi

  KM_DEL="$TMP/key-del.json"
  cat > "$KM_DEL" <<EOF
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-del","actor":{"type":"platform-backend","id":"t"},"params":{"action":"delete","key_ref":"$GEN_KEY_REF"}}
EOF
  run_op "key.manage.delete" "key.manage" "$KM_DEL" 0 "OK" || true

  KM_EXP_AFTER="$TMP/key-exp-after.json"
  cat > "$KM_EXP_AFTER" <<EOF
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-exp-after","actor":{"type":"platform-backend","id":"t"},"params":{"action":"export","key_ref":"$GEN_KEY_REF","allow_plain_export":true}}
EOF
  run_op "key.manage.export.after-delete" "key.manage" "$KM_EXP_AFTER" 6 "KEY_NOT_FOUND" || true
else
  fail "key.manage generate 未返回 key_ref，后续用例无法执行"
fi

KM_BADACT="$TMP/key-badact.json"
cat > "$KM_BADACT" <<'EOF'
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-key-badact","actor":{"type":"platform-backend","id":"t"},"params":{"action":"bogus"}}
EOF
run_op "key.manage.bad-action" "key.manage" "$KM_BADACT" 2 "INVALID_PARAM" || true

# =============================================================
# 4. crypto.service
# =============================================================
echo
echo "---- crypto.service ----"

CS_GEN="$TMP/cs-key-gen.json"
cat > "$CS_GEN" <<'EOF'
{"schema_version":"1.0","operation_id":"key.manage","request_id":"req-p1p2-cs-key-gen","actor":{"type":"platform-backend","id":"t"},"params":{"action":"generate","algorithm":"SM2"}}
EOF
run_op "crypto.setup.key.generate" "key.manage" "$CS_GEN" 0 "OK" || true
CS_KEY_REF="$(jq -r '.data.key_ref // empty' "$TMP/crypto.setup.key.generate.out.json" 2>/dev/null)"

if [ -n "$CS_KEY_REF" ]; then
  ok "crypto.setup key_ref=$CS_KEY_REF"

  CS_SIGN="$TMP/crypto-sign.json"
  cat > "$CS_SIGN" <<EOF
{"schema_version":"1.0","operation_id":"crypto.service","request_id":"req-p1p2-cs-sign","actor":{"type":"platform-backend","id":"t"},"params":{"action":"sign","algorithm":"SM2","key_ref":"$CS_KEY_REF","input_path":"$PAYLOAD_REL","output_path":"tmp/test-p1p2-sig.bin"}}
EOF
  run_op "crypto.service.sign" "crypto.service" "$CS_SIGN" 0 "OK" || true
  CS_SIGN_OUT="$TMP/crypto.service.sign.out.json"
  SIG_REL="$(jq -r '.data.output_path // empty' "$CS_SIGN_OUT" 2>/dev/null)"
  if [ -n "$SIG_REL" ] && [ -s "$CORE_ROOT/$SIG_REL" ]; then
    ok "crypto.service.sign 输出签名文件非空"
  else
    fail "crypto.service.sign 签名文件为空或缺失"
  fi

  CS_TRAV="$TMP/crypto-trav.json"
  cat > "$CS_TRAV" <<EOF
{"schema_version":"1.0","operation_id":"crypto.service","request_id":"req-p1p2-cs-trav","actor":{"type":"platform-backend","id":"t"},"params":{"action":"sign","algorithm":"SM2","key_ref":"$CS_KEY_REF","input_path":"../../../etc/passwd"}}
EOF
  run_op_nonzero "crypto.service.path-traversal" "crypto.service" "$CS_TRAV" 4 || true

  CS_NOKEY="$TMP/crypto-nokey.json"
  cat > "$CS_NOKEY" <<EOF
{"schema_version":"1.0","operation_id":"crypto.service","request_id":"req-p1p2-cs-nokey","actor":{"type":"platform-backend","id":"t"},"params":{"action":"sign","algorithm":"SM2","key_ref":"key-does-not-exist","input_path":"$PAYLOAD_REL"}}
EOF
  run_op "crypto.service.key-not-found" "crypto.service" "$CS_NOKEY" 6 "KEY_NOT_FOUND" || true

  CS_BADACT="$TMP/crypto-badact.json"
  cat > "$CS_BADACT" <<EOF
{"schema_version":"1.0","operation_id":"crypto.service","request_id":"req-p1p2-cs-badact","actor":{"type":"platform-backend","id":"t"},"params":{"action":"bogus","algorithm":"SM2","key_ref":"$CS_KEY_REF","input_path":"$PAYLOAD_REL"}}
EOF
  run_op "crypto.service.bad-action" "crypto.service" "$CS_BADACT" 2 "INVALID_PARAM" || true
else
  skip "crypto.service：无法生成 SM2 密钥"
fi

# =============================================================
# 5. chain.verify
# =============================================================
echo
echo "---- chain.verify ----"
if [ -n "$CA_REL" ]; then
  CH_IN="$TMP/chain.json"
  cat > "$CH_IN" <<EOF
{"schema_version":"1.0","operation_id":"chain.verify","request_id":"req-p1p2-chain","actor":{"type":"platform-backend","id":"t"},"params":{"leaf_path":"$CA_REL","ca_path":"$CA_REL"}}
EOF
  run_op "chain.verify.self" "chain.verify" "$CH_IN" 0 "OK" || true
  CH_OUT="$TMP/chain.verify.self.out.json"
  if [ -f "$CH_OUT" ]; then
    VALID="$(jq -r '.data.valid' "$CH_OUT" 2>/dev/null)"
    [ "$VALID" = "true" ] && ok "chain.verify 自签 CA valid=true" || fail "chain.verify 自签 CA valid=$VALID"
  fi

  if [ -n "$CA_ABS" ] && [ -f "$CA_ABS" ]; then
    sed 's/M/!/g' "$CA_ABS" > "$CORE_ROOT/tmp/test-p1p2-tampered.pem"
    CH_BAD="$TMP/chain-bad.json"
    cat > "$CH_BAD" <<'EOF'
{"schema_version":"1.0","operation_id":"chain.verify","request_id":"req-p1p2-chain-bad","actor":{"type":"platform-backend","id":"t"},"params":{"leaf_path":"tmp/test-p1p2-tampered.pem","ca_path":"tmp/test-p1p2-tampered.pem"}}
EOF
    set +e
    "$DISPATCH" --op chain.verify --in "$CH_BAD" --out "$TMP/chain-bad.out.json" >/dev/null 2>&1
    rc_chbad=$?
    set -e
    if [ "$rc_chbad" -eq 8 ]; then
      ok "chain.verify 篡改证书返回 CERT_PARSE_FAILED"
    elif [ "$rc_chbad" -eq 0 ]; then
      v="$(jq -r '.data.valid' "$TMP/chain-bad.out.json" 2>/dev/null || echo unknown)"
      [ "$v" = "false" ] && ok "chain.verify 篡改证书 valid=false" || fail "chain.verify 篡改证书 valid=$v"
    else
      fail "chain.verify 篡改证书异常 rc=$rc_chbad"
    fi
  fi

  CH_NF="$TMP/chain-nf.json"
  cat > "$CH_NF" <<EOF
{"schema_version":"1.0","operation_id":"chain.verify","request_id":"req-p1p2-chain-nf","actor":{"type":"platform-backend","id":"t"},"params":{"leaf_path":"data/ca/does-not-exist.pem","ca_path":"$CA_REL"}}
EOF
  run_op "chain.verify.not-found" "chain.verify" "$CH_NF" 7 "CERT_NOT_FOUND" || true
else
  skip "chain.verify：无 CA 证书"
fi

# =============================================================
# 6. batch.execute
# =============================================================
echo
echo "---- batch.execute ----"
if [ -n "$CA_REL" ]; then
  BATCH_OK="$TMP/batch.json"
  cat > "$BATCH_OK" <<EOF
{"schema_version":"1.0","operation_id":"batch.execute","request_id":"req-p1p2-batch","actor":{"type":"platform-backend","id":"t"},"params":{"items":[{"operation_id":"cert.parse","params":{"cert_path":"$CA_REL"}},{"operation_id":"cert.parse","params":{"cert_path":"$CA_REL"}}]}}
EOF
  run_op "batch.execute.ok" "batch.execute" "$BATCH_OK" 0 "OK" || true
  BATCH_OK_OUT="$TMP/batch.execute.ok.out.json"
  if [ -f "$BATCH_OK_OUT" ]; then
    TOTAL="$(jq -r '.data.total' "$BATCH_OK_OUT")"
    SC="$(jq -r '.data.success_count' "$BATCH_OK_OUT")"
    FC="$(jq -r '.data.failed_count' "$BATCH_OK_OUT")"
    [ "$TOTAL" = "2" ] && ok "batch.execute total=2"         || fail "batch.execute total=$TOTAL"
    [ "$SC" = "2" ]    && ok "batch.execute success_count=2" || fail "batch.execute success_count=$SC"
    [ "$FC" = "0" ]    && ok "batch.execute failed_count=0"  || fail "batch.execute failed_count=$FC"
  fi

  BATCH_PART="$TMP/batch-part.json"
  cat > "$BATCH_PART" <<EOF
{"schema_version":"1.0","operation_id":"batch.execute","request_id":"req-p1p2-batch-part","actor":{"type":"platform-backend","id":"t"},"params":{"items":[{"operation_id":"cert.parse","params":{"cert_path":"$CA_REL"}},{"operation_id":"cert.parse","params":{"cert_path":"data/ca/does-not-exist.pem"}}]}}
EOF
  run_op "batch.execute.partial-fail" "batch.execute" "$BATCH_PART" 0 "OK" || true
  BATCH_PART_OUT="$TMP/batch.execute.partial-fail.out.json"
  if [ -f "$BATCH_PART_OUT" ]; then
    SC2="$(jq -r '.data.success_count' "$BATCH_PART_OUT")"
    FC2="$(jq -r '.data.failed_count' "$BATCH_PART_OUT")"
    [ "$SC2" = "1" ] && [ "$FC2" = "1" ] && ok "batch.execute 部分失败计数正确 sc=1 fc=1" \
      || fail "batch.execute 部分失败 sc=$SC2 fc=$FC2"
  fi

  BATCH_REC="$TMP/batch-rec.json"
  cat > "$BATCH_REC" <<'EOF'
{"schema_version":"1.0","operation_id":"batch.execute","request_id":"req-p1p2-batch-rec","actor":{"type":"platform-backend","id":"t"},"params":{"items":[{"operation_id":"batch.execute","params":{}}]}}
EOF
  run_op "batch.execute.recursive" "batch.execute" "$BATCH_REC" 2 "INVALID_PARAM" || true

  BATCH_EMPTY="$TMP/batch-empty.json"
  cat > "$BATCH_EMPTY" <<'EOF'
{"schema_version":"1.0","operation_id":"batch.execute","request_id":"req-p1p2-batch-empty","actor":{"type":"platform-backend","id":"t"},"params":{"items":[]}}
EOF
  run_op "batch.execute.empty" "batch.execute" "$BATCH_EMPTY" 2 "INVALID_PARAM" || true
else
  skip "batch.execute：无 CA 证书"
fi

# =============================================================
# 7. ssl.config.generate
# =============================================================
echo
echo "---- ssl.config.generate ----"
if [ -n "$CA_REL" ] && [ -n "$CS_KEY_REF" ]; then
  SSL_IN="$TMP/ssl.json"
  cat > "$SSL_IN" <<EOF
{"schema_version":"1.0","operation_id":"ssl.config.generate","request_id":"req-p1p2-ssl","actor":{"type":"platform-backend","id":"t"},"params":{"server_type":"nginx","cert_path":"$CA_REL","key_ref":"$CS_KEY_REF","output_path":"data/ssl/test-p1p2-nginx.conf"}}
EOF
  run_op "ssl.config.generate.nginx" "ssl.config.generate" "$SSL_IN" 0 "OK" || true
  SSL_OUT="$TMP/ssl.config.generate.nginx.out.json"
  if [ -f "$SSL_OUT" ]; then
    CFG="$(jq -r '.data.config_path // empty' "$SSL_OUT")"
    if [ -n "$CFG" ] && [ -f "$CORE_ROOT/$CFG" ]; then
      ok "ssl.config.generate 输出配置文件存在"
      if grep -q "key_ref:" "$CORE_ROOT/$CFG"; then
        ok "ssl.config.generate 配置使用 key_ref 引用而非私钥"
      else
        fail "ssl.config.generate 配置未使用 key_ref 引用"
      fi
    else
      fail "ssl.config.generate 输出配置文件缺失"
    fi
  fi

  SSL_BAD="$TMP/ssl-bad.json"
  cat > "$SSL_BAD" <<EOF
{"schema_version":"1.0","operation_id":"ssl.config.generate","request_id":"req-p1p2-ssl-bad","actor":{"type":"platform-backend","id":"t"},"params":{"server_type":"apache","cert_path":"$CA_REL","key_ref":"$CS_KEY_REF"}}
EOF
  run_op "ssl.config.generate.bad-server" "ssl.config.generate" "$SSL_BAD" 2 "INVALID_PARAM" || true
else
  skip "ssl.config.generate：缺 CA 或 key_ref"
fi

# =============================================================
# 8. pqc.cert.create
# =============================================================
echo
echo "---- pqc.cert.create ----"

# ML-DSA 软断言
PQC_IN="$TMP/pqc.json"
cat > "$PQC_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"pqc.cert.create","request_id":"req-p1p2-pqc","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"ML-DSA","key_params":{"parameter":"ML-DSA-65"},"subject":{"CN":"pqc.example.com","O":"Test"},"validity_days":365,"cert_type":"server"}}
EOF
set +e
"$DISPATCH" --op pqc.cert.create --in "$PQC_IN" --out "$TMP/pqc.out.json" >/dev/null 2>&1
rc_pqc=$?
set -e
PQC_CODE="$(jq -r '.code // empty' "$TMP/pqc.out.json" 2>/dev/null || echo "")"
case "$rc_pqc:$PQC_CODE" in
  0:OK)
    ok "pqc.cert.create ML-DSA 成功（铜锁支持）"
    ;;
  12:VERSION_UNSUPPORTED)
    ok "pqc.cert.create ML-DSA 返回 VERSION_UNSUPPORTED（铜锁 pre 版不支持，可接受）"
    ;;
  3:ALGORITHM_NOT_ALLOWED)
    ok "pqc.cert.create ML-DSA 返回 ALGORITHM_NOT_ALLOWED（白名单未开启，可接受）"
    ;;
  *)
    fail "pqc.cert.create ML-DSA 异常 rc=$rc_pqc code=$PQC_CODE"
    ;;
esac

# ML-KEM 用于签名证书应被拒绝
PQC_BAD_IN="$TMP/pqc-bad.json"
cat > "$PQC_BAD_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"pqc.cert.create","request_id":"req-p1p2-pqc-bad","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"ML-KEM","subject":{"CN":"bad.example.com"},"validity_days":365,"cert_type":"server"}}
EOF
set +e
"$DISPATCH" --op pqc.cert.create --in "$PQC_BAD_IN" --out "$TMP/pqc-bad.out.json" >/dev/null 2>&1
rc_pqcbad=$?
set -e
PQC_BAD_CODE="$(jq -r '.code // empty' "$TMP/pqc-bad.out.json" 2>/dev/null || echo "")"
case "$rc_pqcbad:$PQC_BAD_CODE" in
  3:ALGORITHM_NOT_ALLOWED)
    ok "pqc.cert.create ML-KEM 用于签名证书被拒绝"
    ;;
  12:VERSION_UNSUPPORTED)
    ok "pqc.cert.create ML-KEM 返回 VERSION_UNSUPPORTED（铜锁不支持）"
    ;;
  *)
    fail "pqc.cert.create ML-KEM 异常 rc=$rc_pqcbad code=$PQC_BAD_CODE"
    ;;
esac

PQC_BADALG="$TMP/pqc-badalg.json"
cat > "$PQC_BADALG" <<'EOF'
{"schema_version":"1.0","operation_id":"pqc.cert.create","request_id":"req-p1p2-pqc-badalg","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"DES","subject":{"CN":"x"},"validity_days":365}}
EOF
run_op "pqc.cert.create.bad-alg" "pqc.cert.create" "$PQC_BADALG" 3 "ALGORITHM_NOT_ALLOWED" || true

# =============================================================
# 9. 审计脱敏（只扫描本轮新增条目）
# =============================================================
echo
echo "---- 审计脱敏 ----"
if [ -f "$AUDIT_LOG" ]; then

  # 至少一条 cert.parse 审计，且 params_digest 以 sha256: 开头
  cnt="$(audit_tail | jq -r 'select(
      .request_id=="req-p1p2-parse"
      and ((.params_digest // "") | startswith("sha256:"))
    ) | 1' 2>/dev/null | wc -l | tr -d ' ')"
  if [ "$cnt" -gt 0 ]; then
    ok "cert.parse 审计 params_digest 格式正确（$cnt 条）"
  else
    fail "cert.parse 审计 params_digest 缺失或格式异常"
  fi

  # 至少一条 crypto.service 成功审计
  cnt="$(audit_tail | jq -r 'select(
      .request_id=="req-p1p2-cs-sign"
      and .operation_id=="crypto.service"
      and .result=="SUCCESS"
    ) | 1' 2>/dev/null | wc -l | tr -d ' ')"
  if [ "$cnt" -gt 0 ]; then
    ok "crypto.service.sign 审计已写入（$cnt 条）"
  else
    fail "crypto.service.sign 审计未写入"
  fi

  # 至少一条 key.manage 审计
  cnt="$(audit_tail | jq -r 'select(
      .operation_id=="key.manage"
      and .result=="SUCCESS"
    ) | 1' 2>/dev/null | wc -l | tr -d ' ')"
  if [ "$cnt" -gt 0 ]; then
    ok "key.manage 审计已写入（$cnt 条）"
  else
    fail "key.manage 审计未写入"
  fi

  # 审计不含私钥或口令
  if audit_tail | grep -E 'BEGIN.*PRIVATE KEY|test-password-12345' >/dev/null 2>&1; then
    fail "审计日志泄露私钥或口令"
  else
    ok "审计日志无私钥/口令泄露"
  fi
else
  fail "审计日志不存在: $AUDIT_LOG"
fi

# =============================================================
# 10. 安全基线
# =============================================================
echo
echo "---- 安全基线 ----"
hits="$(grep -RIn --exclude-dir=.git -E '\bshell=True\b|\beval\b|\bsh -c\b' \
        "$CORE_ROOT/sbin" "$CORE_ROOT/src/tool/src" 2>/dev/null || true)"
if [ -n "$hits" ]; then
  fail "发现 shell=True/eval/sh -c：$hits"
else
  ok "无 shell=True/eval/sh -c"
fi

# =============================================================
# 11. 清理测试临时文件
# =============================================================
rm -f "$CORE_ROOT/tmp/test-p1p2-payload.txt" \
      "$CORE_ROOT/tmp/test-p1p2-pass.txt" \
      "$CORE_ROOT/tmp/test-p1p2-badpass.txt" \
      "$CORE_ROOT/tmp/test-p1p2-bad.pem" \
      "$CORE_ROOT/tmp/test-p1p2-tampered.pem" \
      "$CORE_ROOT/tmp/test-p1p2-sig.bin" \
      2>/dev/null || true

# =============================================================
# 汇总
# =============================================================
echo
echo "============================================"
echo " SUMMARY: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"
echo "============================================"
[ "$FAIL" -eq 0 ] || exit 1
exit 0
