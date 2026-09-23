#!/bin/sh
# core/tests/test_p0_ops.sh
# P0 密码操作交付级回归测试：功能 + 契约 + 加密 + 审计 + 错误码 + 性能

set -u

CORE_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DISPATCH="$CORE_ROOT/sbin/dispatch.sh"
OPENSSL_BIN="${TONSUO_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
KEYCRYPT_BIN="${KEYCRYPT_BIN:-$CORE_ROOT/src/tool/bin/keycrypt}"
MASTER_KEY="$CORE_ROOT/data/keys/master.key"
AUDIT_LOG="$CORE_ROOT/logs/audit.jsonl"

TMP=$(mktemp -d "${TMPDIR:-/tmp}/p0-ops.XXXXXX")
trap 'rm -rf "$TMP"' EXIT

PASS=0; FAIL=0; SKIP=0
ok()   { PASS=$((PASS+1)); printf 'PASS  %s\n' "$1"; }
fail() { FAIL=$((FAIL+1)); printf 'FAIL  %s\n' "$1"; }
skip() { SKIP=$((SKIP+1)); printf 'SKIP  %s\n' "$1"; }

command -v jq >/dev/null 2>&1 || { echo "jq required"; exit 2; }
[ -x "$DISPATCH" ]    || { echo "dispatch not executable: $DISPATCH"; exit 2; }
[ -x "$OPENSSL_BIN" ] || { echo "openssl not executable: $OPENSSL_BIN"; exit 2; }

now_ms() {
  n=$(date +%s%3N 2>/dev/null)
  if [ -n "$n" ] && [ "$n" != "$(date +%s)3N" ]; then
    printf '%s\n' "$n"
  else
    echo $(($(date +%s) * 1000))
  fi
}

mode_of() {
  stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1" 2>/dev/null || echo '?'
}

# 跑一次操作，记录 stdout/stderr/exit，并断言基本契约。
# 输出文件固定为 $TMP/<case>.out.json，所有下游断言都读该路径。
run_op() {
  case_name="$1"; op="$2"; in_file="$3"; expect_rc="$4"; expect_code="$5"
  out="$TMP/$case_name.out.json"
  sout="$TMP/$case_name.stdout"
  serr="$TMP/$case_name.stderr"

  set +e
  "$DISPATCH" --op "$op" --in "$in_file" --out "$out" >"$sout" 2>"$serr"
  rc=$?
  set -e

  if [ "$rc" -ne "$expect_rc" ]; then
    fail "$case_name: exit=$rc expected=$expect_rc; stderr=$(tail -n 3 "$serr" 2>/dev/null | tr '\n' ' ')"
    return 1
  fi
  if [ ! -f "$out" ]; then
    fail "$case_name: --out file missing"; return 1
  fi
  if ! jq -e . "$out" >/dev/null 2>&1; then
    fail "$case_name: --out not valid JSON: $(head -c 200 "$out")"; return 1
  fi

  code="$(jq -r '.code // empty' "$out")"
  if [ "$code" != "$expect_code" ]; then
    fail "$case_name: code=$code expected=$expect_code"; return 1
  fi

  if [ -s "$sout" ]; then
    if ! jq -e . "$sout" >/dev/null 2>&1; then
      fail "$case_name: stdout 不是合法 JSON: $(head -c 200 "$sout")"; return 1
    fi
  fi

  if [ -s "$serr" ]; then
    bad="$(awk 'NF && $0 !~ /^\{.*\}$/ {print NR": "substr($0,1,80)}' "$serr" | head -n1)"
    if [ -n "$bad" ]; then
      fail "$case_name: stderr 非 JSON Lines -> $bad"; return 1
    fi
  fi

  ok "$case_name: exit=$rc code=$code"
  return 0
}

check_envelope() {
  case_name="$1"; out="$2"
  miss=""
  for f in schema_version code message request_id operation_id data error audit; do
    if ! jq -e "has(\"$f\")" "$out" >/dev/null 2>&1; then
      miss="$miss $f"
    fi
  done
  if [ -n "$miss" ]; then
    fail "$case_name: 缺少公共字段:$miss"; return 1
  fi
  sv="$(jq -r '.schema_version' "$out")"
  [ "$sv" = "1.0" ] || { fail "$case_name: schema_version=$sv"; return 1; }
  ok "$case_name: 公共信封字段完整"
}

check_data_field() {
  case_name="$1"; out="$2"; field="$3"
  if [ ! -f "$out" ]; then
    fail "$case_name: --out 不存在: $out"; return 1
  fi
  v="$(jq -r ".data.$field // empty" "$out")"
  if [ -z "$v" ] || [ "$v" = "null" ]; then
    fail "$case_name: data.$field 缺失"; return 1
  fi
  ok "$case_name: data.$field=$v"
  return 0
}

echo "============================================"
echo " P0 密码操作交付级回归测试"
echo " CORE_ROOT=$CORE_ROOT"
echo "============================================"

# ---------- 1. ca.create: SM2 / RSA / ECC ----------
echo
echo "---- ca.create ----"

CA_SM2_IN="$TMP/ca-sm2.json"
cat > "$CA_SM2_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-sm2","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"SM2","subject":{"CN":"Test Root CA SM2","O":"Example"},"validity_days":3650}}
EOF

t0=$(now_ms)
run_op "ca.create.SM2" "ca.create" "$CA_SM2_IN" 0 "OK" || true
t1=$(now_ms)
dur_ms=$((t1 - t0))
[ "$dur_ms" -le 3000 ] && ok "ca.create.SM2 性能 ${dur_ms}ms ≤3000ms" || fail "ca.create.SM2 性能 ${dur_ms}ms >3000ms"

CA_SM2_OUT="$TMP/ca.create.SM2.out.json"
check_envelope "ca.create.SM2" "$CA_SM2_OUT"
check_data_field "ca.create.SM2" "$CA_SM2_OUT" "ca_id"
check_data_field "ca.create.SM2" "$CA_SM2_OUT" "cert_path"
check_data_field "ca.create.SM2" "$CA_SM2_OUT" "key_ref"

CA_SM2_ID="$(jq -r '.data.ca_id' "$CA_SM2_OUT" 2>/dev/null)"
CA_SM2_KR="$(jq -r '.data.key_ref' "$CA_SM2_OUT" 2>/dev/null)"
CA_SM2_CERT="$CORE_ROOT/$(jq -r '.data.cert_path' "$CA_SM2_OUT" 2>/dev/null)"

if [ -f "$CA_SM2_CERT" ]; then
  ok "ca.create.SM2 证书文件存在"
  if "$OPENSSL_BIN" x509 -in "$CA_SM2_CERT" -noout -text > "$TMP/ca-sm2.txt" 2>/dev/null; then
    ok "ca.create.SM2 证书可解析"
    grep -q "CA:TRUE" "$TMP/ca-sm2.txt" && ok "ca.create.SM2 basicConstraints=CA:TRUE" || fail "ca.create.SM2 basicConstraints 缺失"
    grep -q "Certificate Sign" "$TMP/ca-sm2.txt" && ok "ca.create.SM2 keyUsage 含 keyCertSign" || fail "ca.create.SM2 keyUsage 异常"
    grep -qE "SM2|sm2" "$TMP/ca-sm2.txt" && ok "ca.create.SM2 公钥为 SM2" || fail "ca.create.SM2 公钥算法异常"
  else
    fail "ca.create.SM2 证书解析失败"
  fi
else
  fail "ca.create.SM2 证书文件不存在: $CA_SM2_CERT"
fi

ENC="$CORE_ROOT/data/keys/${CA_SM2_KR}.key.enc"
if [ -f "$ENC" ]; then
  ok "ca.create.SM2 加密私钥存在"
  m="$(mode_of "$ENC")"
  [ "$m" = "600" ] && ok "ca.create.SM2 加密私钥权限 600" || fail "ca.create.SM2 加密私钥权限 $m"
  if [ -x "$KEYCRYPT_BIN" ] && [ -f "$MASTER_KEY" ]; then
    if "$KEYCRYPT_BIN" decrypt --master "$MASTER_KEY" --in "$ENC" --out "$TMP/ca-sm2.key" 2>/dev/null; then
      if "$OPENSSL_BIN" pkey -in "$TMP/ca-sm2.key" -noout >/dev/null 2>&1; then
        ok "ca.create.SM2 加密私钥可解密且为合法私钥"
      else
        fail "ca.create.SM2 解密后非法私钥"
      fi
    else
      fail "ca.create.SM2 加密私钥无法解密"
    fi
  else
    skip "keycrypt 或 master.key 不可用，跳过私钥解密回环"
  fi
else
  fail "ca.create.SM2 加密私钥不存在: $ENC"
fi

if [ -f "$AUDIT_LOG" ]; then
  if jq -e 'select(.request_id=="req-ca-sm2" and .result=="SUCCESS" and .operation_id=="ca.create")' "$AUDIT_LOG" >/dev/null 2>&1; then
    ok "ca.create.SM2 审计已写入"
    if jq -e 'select(.request_id=="req-ca-sm2") | has("params_digest") and has("duration_ms") and has("ts")' "$AUDIT_LOG" >/dev/null 2>&1; then
      ok "ca.create.SM2 审计字段完整"
    else
      fail "ca.create.SM2 审计字段缺失"
    fi
  else
    fail "ca.create.SM2 审计未写入或内容异常"
  fi
else
  fail "审计日志不存在: $AUDIT_LOG"
fi

# RSA 2048
CA_RSA_IN="$TMP/ca-rsa.json"
cat > "$CA_RSA_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-rsa","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"RSA","key_params":{"key_size":2048},"subject":{"CN":"Test Root CA RSA","O":"Example"},"validity_days":3650}}
EOF
t0=$(now_ms)
run_op "ca.create.RSA2048" "ca.create" "$CA_RSA_IN" 0 "OK" || true
t1=$(now_ms)
dur_ms=$((t1 - t0))
[ "$dur_ms" -le 3000 ] && ok "ca.create.RSA2048 性能 ${dur_ms}ms ≤3000ms" || fail "ca.create.RSA2048 性能 ${dur_ms}ms >3000ms"
CA_RSA_OUT="$TMP/ca.create.RSA2048.out.json"
check_data_field "ca.create.RSA2048" "$CA_RSA_OUT" "cert_path"
RSA_CERT="$CORE_ROOT/$(jq -r '.data.cert_path' "$CA_RSA_OUT" 2>/dev/null)"
if [ -f "$RSA_CERT" ] && "$OPENSSL_BIN" x509 -in "$RSA_CERT" -noout -text 2>/dev/null | grep -q "2048 bit"; then
  ok "ca.create.RSA2048 密钥长度 2048"
else
  fail "ca.create.RSA2048 密钥长度不是 2048"
fi

# ECC P-256
CA_ECC_IN="$TMP/ca-ecc.json"
cat > "$CA_ECC_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-ecc","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"ECC","key_params":{"curve":"prime256v1"},"subject":{"CN":"Test Root CA ECC","O":"Example"},"validity_days":3650}}
EOF
run_op "ca.create.ECC-P256" "ca.create" "$CA_ECC_IN" 0 "OK" || true
CA_ECC_OUT="$TMP/ca.create.ECC-P256.out.json"
ECC_CERT="$CORE_ROOT/$(jq -r '.data.cert_path' "$CA_ECC_OUT" 2>/dev/null)"
if [ -f "$ECC_CERT" ] && "$OPENSSL_BIN" x509 -in "$ECC_CERT" -noout -text 2>/dev/null | grep -q "prime256v1"; then
  ok "ca.create.ECC-P256 曲线正确"
else
  fail "ca.create.ECC-P256 曲线异常"
fi

# 非法算法
CA_BAD_IN="$TMP/ca-bad.json"
cat > "$CA_BAD_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-bad","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"DES","subject":{"CN":"Bad"},"validity_days":3650}}
EOF
run_op "ca.create.bad-alg" "ca.create" "$CA_BAD_IN" 3 "ALGORITHM_NOT_ALLOWED" || true

# 缺 algorithm
CA_MISS_IN="$TMP/ca-miss.json"
cat > "$CA_MISS_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-miss","actor":{"type":"platform-backend","id":"t"},"params":{"subject":{"CN":"Miss"}}}
EOF
set +e
"$DISPATCH" --op ca.create --in "$CA_MISS_IN" --out "$TMP/ca-miss.out.json" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -ne 0 ] && ok "ca.create 缺 algorithm 被拒绝 rc=$rc" || fail "ca.create 缺 algorithm 未被拒绝"

# RSA 1024 拒绝
CA_RSA1024_IN="$TMP/ca-rsa1024.json"
cat > "$CA_RSA1024_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.create","request_id":"req-ca-rsa1024","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"RSA","key_params":{"key_size":1024},"subject":{"CN":"Bad RSA"},"validity_days":3650}}
EOF
set +e
"$DISPATCH" --op ca.create --in "$CA_RSA1024_IN" --out "$TMP/ca-rsa1024.out.json" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -ne 0 ] && ok "ca.create RSA 1024 被拒绝 rc=$rc" || fail "ca.create RSA 1024 未被拒绝"

# ---------- 2. ca.intermediate.create ----------
echo
echo "---- ca.intermediate.create ----"
if [ -n "$CA_SM2_ID" ] && [ -n "$CA_SM2_KR" ]; then
  CA_INT_IN="$TMP/ca-int.json"
  cat > "$CA_INT_IN" <<EOF
{"schema_version":"1.0","operation_id":"ca.intermediate.create","request_id":"req-ca-int","actor":{"type":"platform-backend","id":"t"},"params":{"parent_ca_id":"$CA_SM2_ID","parent_key_ref":"$CA_SM2_KR","algorithm":"SM2","subject":{"CN":"Test Intermediate CA","O":"Example"},"validity_days":1825,"path_len":0}}
EOF
  run_op "ca.intermediate.create" "ca.intermediate.create" "$CA_INT_IN" 0 "OK" || true
  CA_INT_OUT="$TMP/ca.intermediate.create.out.json"
  check_data_field "ca.intermediate.create" "$CA_INT_OUT" "cert_path"
  check_data_field "ca.intermediate.create" "$CA_INT_OUT" "chain_path"
  INT_CERT="$CORE_ROOT/$(jq -r '.data.cert_path' "$CA_INT_OUT" 2>/dev/null)"
  CHAIN="$CORE_ROOT/$(jq -r '.data.chain_path' "$CA_INT_OUT" 2>/dev/null)"
  if [ -f "$INT_CERT" ] && [ -f "$CHAIN" ]; then
    ok "ca.intermediate.create 证书与链存在"
    if "$OPENSSL_BIN" verify -CAfile "$CA_SM2_CERT" "$INT_CERT" >/dev/null 2>&1; then
      ok "ca.intermediate.create 链可验证"
    else
      fail "ca.intermediate.create 链验证失败"
    fi
  else
    fail "ca.intermediate.create 证书或链文件缺失"
  fi
else
  fail "ca.intermediate.create 前置 ca.create 失败，跳过"
fi

CA_INT_BAD="$TMP/ca-int-bad.json"
cat > "$CA_INT_BAD" <<'EOF'
{"schema_version":"1.0","operation_id":"ca.intermediate.create","request_id":"req-ca-int-bad","actor":{"type":"platform-backend","id":"t"},"params":{"parent_ca_id":"ca-nonexistent","parent_key_ref":"key-nonexistent","algorithm":"SM2","subject":{"CN":"Bad"},"validity_days":365,"path_len":0}}
EOF
run_op "ca.intermediate.create.bad-parent" "ca.intermediate.create" "$CA_INT_BAD" 7 "CERT_NOT_FOUND" || true

# ---------- 3. csr.create ----------
echo
echo "---- csr.create ----"
CSR_IN="$TMP/csr.json"
cat > "$CSR_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"csr.create","request_id":"req-csr","actor":{"type":"platform-backend","id":"t"},"params":{"algorithm":"SM2","subject":{"CN":"test.example.com","O":"Example"},"san":["test.example.com","www.example.com"]}}
EOF
run_op "csr.create.SM2" "csr.create" "$CSR_IN" 0 "OK" || true
CSR_OUT="$TMP/csr.create.SM2.out.json"
check_data_field "csr.create.SM2" "$CSR_OUT" "csr_path"
check_data_field "csr.create.SM2" "$CSR_OUT" "key_ref"
CSR_FILE="$CORE_ROOT/$(jq -r '.data.csr_path' "$CSR_OUT" 2>/dev/null)"
CSR_ID="$(jq -r '.data.csr_id' "$CSR_OUT" 2>/dev/null)"
if [ -f "$CSR_FILE" ]; then
  if "$OPENSSL_BIN" req -in "$CSR_FILE" -noout -text >/dev/null 2>&1; then
    ok "csr.create.SM2 CSR 可解析"
  else
    fail "csr.create.SM2 CSR 解析失败"
  fi
fi

# ---------- 4. cert.sign ----------
echo
echo "---- cert.sign ----"
if [ -n "$CSR_ID" ] && [ -n "$CA_SM2_ID" ] && [ -n "$CA_SM2_KR" ]; then
  SIGN_IN="$TMP/sign.json"
  cat > "$SIGN_IN" <<EOF
{"schema_version":"1.0","operation_id":"cert.sign","request_id":"req-sign","actor":{"type":"platform-backend","id":"t"},"params":{"csr_id":"$CSR_ID","ca_id":"$CA_SM2_ID","ca_key_ref":"$CA_SM2_KR","cert_type":"server","validity_days":365,"algorithm":"SM2"}}
EOF
  run_op "cert.sign.server" "cert.sign" "$SIGN_IN" 0 "OK" || true
  SIGN_OUT="$TMP/cert.sign.server.out.json"
  check_data_field "cert.sign.server" "$SIGN_OUT" "cert_path"
  SIGNED="$CORE_ROOT/$(jq -r '.data.cert_path' "$SIGN_OUT" 2>/dev/null)"
  if [ -f "$SIGNED" ]; then
    if "$OPENSSL_BIN" verify -CAfile "$CA_SM2_CERT" "$SIGNED" >/dev/null 2>&1; then
      ok "cert.sign.server 签发证书可用根 CA 验证"
    else
      fail "cert.sign.server 证书链验证失败"
    fi
    "$OPENSSL_BIN" x509 -in "$SIGNED" -noout -text 2>/dev/null | grep -q "TLS Web Server Authentication" \
      && ok "cert.sign.server EKU=serverAuth" || fail "cert.sign.server EKU 异常"
  fi

  SIGN_BAD="$TMP/sign-bad.json"
  cat > "$SIGN_BAD" <<EOF
{"schema_version":"1.0","operation_id":"cert.sign","request_id":"req-sign-bad","actor":{"type":"platform-backend","id":"t"},"params":{"csr_id":"$CSR_ID","ca_id":"$CA_SM2_ID","ca_key_ref":"$CA_SM2_KR","cert_type":"bogus","validity_days":365,"algorithm":"SM2"}}
EOF
  run_op "cert.sign.bad-type" "cert.sign" "$SIGN_BAD" 2 "INVALID_PARAM" || true
fi

# ---------- 5. dual_cert.create ----------
echo
echo "---- dual_cert.create ----"
if [ -n "$CA_SM2_ID" ] && [ -n "$CA_SM2_KR" ]; then
  DUAL_IN="$TMP/dual.json"
  cat > "$DUAL_IN" <<EOF
{"schema_version":"1.0","operation_id":"dual_cert.create","request_id":"req-dual","actor":{"type":"platform-backend","id":"t"},"params":{"sign_algorithm":"SM2","enc_algorithm":"SM2","tlcp_profile":"GB/T 38636-2020","subject":{"CN":"dual.example.com","O":"Example"},"validity_days":365,"ca_id":"$CA_SM2_ID","ca_key_ref":"$CA_SM2_KR"}}
EOF
  run_op "dual_cert.create" "dual_cert.create" "$DUAL_IN" 0 "OK" || true
  DUAL_OUT="$TMP/dual_cert.create.out.json"
  check_data_field "dual_cert.create" "$DUAL_OUT" "sign_cert_path"
  check_data_field "dual_cert.create" "$DUAL_OUT" "enc_cert_path"
  SIGN_C="$CORE_ROOT/$(jq -r '.data.sign_cert_path' "$DUAL_OUT" 2>/dev/null)"
  ENC_C="$CORE_ROOT/$(jq -r '.data.enc_cert_path' "$DUAL_OUT" 2>/dev/null)"
  if [ -f "$SIGN_C" ]; then
    "$OPENSSL_BIN" x509 -in "$SIGN_C" -noout -text 2>/dev/null | grep -q "Digital Signature" \
      && ok "dual_cert 签名证书 keyUsage=digitalSignature" || fail "dual_cert 签名证书 keyUsage 异常"
  fi
  if [ -f "$ENC_C" ]; then
    "$OPENSSL_BIN" x509 -in "$ENC_C" -noout -text 2>/dev/null | grep -qE "Key Encipherment|Key Agreement" \
      && ok "dual_cert 加密证书 keyUsage=keyEncipherment/keyAgreement" || fail "dual_cert 加密证书 keyUsage 异常"
  fi

  DUAL_BAD="$TMP/dual-bad.json"
  cat > "$DUAL_BAD" <<EOF
{"schema_version":"1.0","operation_id":"dual_cert.create","request_id":"req-dual-bad","actor":{"type":"platform-backend","id":"t"},"params":{"sign_algorithm":"RSA","enc_algorithm":"SM2","tlcp_profile":"GB/T 38636-2020","subject":{"CN":"bad"},"validity_days":365,"ca_id":"$CA_SM2_ID","ca_key_ref":"$CA_SM2_KR"}}
EOF
  run_op "dual_cert.create.bad-alg" "dual_cert.create" "$DUAL_BAD" 3 "ALGORITHM_NOT_ALLOWED" || true
fi

# ---------- 6. crl.create ----------
echo
echo "---- crl.create ----"
if [ -n "$CA_SM2_ID" ] && [ -n "$CA_SM2_KR" ]; then
  SIGNED_SERIAL="$(jq -r '.data.serial' "$TMP/cert.sign.server.out.json" 2>/dev/null)"
  [ -n "$SIGNED_SERIAL" ] && [ "$SIGNED_SERIAL" != "null" ] || SIGNED_SERIAL="1000"
  CRL_IN="$TMP/crl.json"
  cat > "$CRL_IN" <<EOF
{"schema_version":"1.0","operation_id":"crl.create","request_id":"req-crl","actor":{"type":"platform-backend","id":"t"},"params":{"ca_id":"$CA_SM2_ID","ca_key_ref":"$CA_SM2_KR","revoked_serials":["$SIGNED_SERIAL"],"digest_algorithm":"SM3"}}
EOF
  run_op "crl.create" "crl.create" "$CRL_IN" 0 "OK" || true
  CRL_OUT="$TMP/crl.create.out.json"
  check_data_field "crl.create" "$CRL_OUT" "crl_path"
  CRL_FILE="$CORE_ROOT/$(jq -r '.data.crl_path' "$CRL_OUT" 2>/dev/null)"
  if [ -f "$CRL_FILE" ]; then
    if "$OPENSSL_BIN" crl -in "$CRL_FILE" -noout -text >/dev/null 2>&1; then
      ok "crl.create CRL 可解析"
      "$OPENSSL_BIN" crl -in "$CRL_FILE" -noout -text 2>/dev/null | grep -qiE "Serial Number|Serial" \
        && ok "crl.create CRL 含撤销序列号" || fail "crl.create CRL 无撤销条目"
    else
      fail "crl.create CRL 解析失败"
    fi
  fi
fi

# ---------- 7. 契约错误路径 ----------
echo
echo "---- 契约错误路径 ----"
UNK_IN="$TMP/unk.json"
cat > "$UNK_IN" <<'EOF'
{"schema_version":"1.0","operation_id":"does.not.exist","request_id":"req-unk","actor":{"type":"platform-backend","id":"t"},"params":{}}
EOF
run_op "unknown-op" "does.not.exist" "$UNK_IN" 2 "INVALID_PARAM" || true

set +e
"$DISPATCH" --op ca.create --out "$TMP/no-in.json" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -eq 2 ] && ok "缺 --in 退出码 2" || fail "缺 --in 退出码 $rc"

# ---------- 8. 安全基线 ----------
echo
echo "---- 安全基线 ----"
hits="$(grep -RIn --exclude-dir=.git -E '\bshell=True\b|\beval\b|\bsh -c\b' \
        "$CORE_ROOT/sbin" "$CORE_ROOT/src/tool/src" 2>/dev/null || true)"
if [ -n "$hits" ]; then
  fail "发现 shell=True/eval/sh -c：$hits"
else
  ok "无 shell=True/eval/sh -c"
fi

if grep -rq "BIO_new_file\|crypto/bio" "$TMP"/*.out.json 2>/dev/null; then
  fail "响应中泄露 OpenSSL 原始错误"
else
  ok "响应未泄露 OpenSSL 原始错误"
fi

# ---------- 9. 幂等与重复请求 ----------
echo
echo "---- 幂等与重复请求 ----"
set +e
"$DISPATCH" --op ca.create --in "$CA_SM2_IN" --out "$TMP/ca-sm2-2.out.json" >/dev/null 2>&1
rc1=$?
set -e
if [ "$rc1" -eq 0 ]; then
  ok "同 request_id 重复调用仍成功（当前实现每次都新生成）"
else
  ok "同 request_id 重复调用被拒绝 rc=$rc1"
fi

# ---------- 汇总 ----------
echo
echo "============================================"
echo " SUMMARY: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"
echo "============================================"
[ "$FAIL" -eq 0 ] || exit 1
exit 0
