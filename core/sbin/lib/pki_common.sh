#!/bin/sh
# core/sbin/lib/pki_common.sh
# P0 密码操作公共函数。所有函数不依赖 bash 特性，兼容 dash/POSIX sh。

_CORE_PKI_LOADED="${_CORE_PKI_LOADED:-0}"
if [ "$_CORE_PKI_LOADED" = "1" ]; then
  return 0 2>/dev/null || true
fi
_CORE_PKI_LOADED=1

CORE_ROOT="${CORE_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)}"
export CORE_ROOT
OPENSSL_BIN="${TONSUO_OPENSSL_BIN:-$CORE_ROOT/libs/bin/tongsuo/bin/openssl}"
export OPENSSL_BIN

# 优先使用项目自管 openssl.cnf，其次编译产物，最后留给 OpenSSL 自身处理
if [ -z "${OPENSSL_CONF:-}" ]; then
  if [ -f "$CORE_ROOT/conf/openssl.cnf" ]; then
    OPENSSL_CONF="$CORE_ROOT/conf/openssl.cnf"
  elif [ -f "$CORE_ROOT/libs/bin/tongsuo/ssl/openssl.cnf" ]; then
    OPENSSL_CONF="$CORE_ROOT/libs/bin/tongsuo/ssl/openssl.cnf"
  fi
  export OPENSSL_CONF
fi

pki_require_jq() {
  command -v jq >/dev/null 2>&1 || {
    printf '{"schema_version":"1.0","code":"INTERNAL_ERROR","message":"jq not found","data":null,"error":{"code":"INTERNAL_ERROR","message":"jq not found","detail":{},"retryable":false}}\n' >&2
    exit 99
  }
}

json_get() {
  file="$1"; path="$2"
  jq -r "$path // empty" "$file" 2>/dev/null || true
}

json_get_def() {
  file="$1"; path="$2"; def="$3"
  v="$(jq -r "$path // empty" "$file" 2>/dev/null || true)"
  if [ -n "$v" ]; then
    printf '%s\n' "$v"
  else
    printf '%s\n' "$def"
  fi
}

pki_new_id() {
  prefix="$1"
  rnd="$(head -c 4 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  printf '%s-%s-%s-%s\n' "$prefix" "$(date -u +%Y%m%d%H%M%S)" "$$" "$rnd"
}

pki_new_key_ref() {
  pki_new_id "key"
}

pki_rel_path() {
  p="$1"
  case "$p" in
    /*) printf '%s\n' "$p" ;;
    *)  printf '%s\n' "$CORE_ROOT/$p" ;;
  esac
}

pki_mkdir_for() {
  d="$(dirname "$1")"
  [ -d "$d" ] || mkdir -p "$d"
}

# 统一错误输出。detail 若为空或非法 JSON，则降级为 {}。
pki_fail() {
  out="$1"; code="$2"; msg="$3"; detail="${4:-}"; exit_code="${5:-2}"

  if [ -z "$detail" ] || ! printf '%s' "$detail" | jq -e . >/dev/null 2>&1; then
    detail='{}'
  fi

  if [ -z "$out" ]; then
    printf '{"schema_version":"1.0","code":"%s","message":"%s","request_id":null,"operation_id":null,"task_id":null,"data":null,"error":{"code":"%s","message":"%s","detail":%s,"retryable":false},"audit":null}\n' \
      "$code" "$msg" "$code" "$msg" "$detail" >&2
    exit "$exit_code"
  fi

  d="$(dirname "$out")"
  [ -d "$d" ] || mkdir -p "$d"

  if ! jq -n \
      --arg code "$code" \
      --arg msg  "$msg" \
      --argjson detail "$detail" \
      '{schema_version:"1.0",code:$code,message:$msg,request_id:null,operation_id:null,task_id:null,data:null,error:{code:$code,message:$msg,detail:$detail,retryable:false},audit:null}' \
      > "$out" 2>/dev/null; then
    printf '{"schema_version":"1.0","code":"%s","message":"%s","request_id":null,"operation_id":null,"task_id":null,"data":null,"error":{"code":"%s","message":"%s","detail":{},"retryable":false},"audit":null}\n' \
      "$code" "$msg" "$code" "$msg" > "$out"
  fi

  exit "$exit_code"
}

pki_success() {
  in="$1"; out="$2"; op="$3"; req="$4"; task="$5"; data_json="$6"; audit_json="$7"
  if [ -z "$data_json" ] || ! printf '%s' "$data_json" | jq -e . >/dev/null 2>&1; then
    data_json='{}'
  fi
  if [ -z "$audit_json" ] || ! printf '%s' "$audit_json" | jq -e . >/dev/null 2>&1; then
    audit_json='{}'
  fi

  jq -n \
    --arg op "$op" \
    --arg req "$req" \
    --arg task "$task" \
    --argjson data "$data_json" \
    --argjson audit "$audit_json" \
    '{schema_version:"1.0",code:"OK",message:"success",request_id:(if $req=="" then null else $req end),operation_id:$op,task_id:(if $task=="" then null else $task end),data:$data,error:null,audit:$audit}' \
    > "$out"
}

# 每个 operation 允许的算法白名单
pki_check_algorithm() {
  op="$1"; alg="$2"
  case "$op:$alg" in
    ca.create:SM2|ca.create:RSA|ca.create:ECC|ca.create:ML-DSA) return 0 ;;
    ca.intermediate.create:SM2|ca.intermediate.create:RSA|ca.intermediate.create:ECC|ca.intermediate.create:ML-DSA) return 0 ;;
    csr.create:SM2|csr.create:RSA|csr.create:ECC|csr.create:ML-DSA) return 0 ;;
    cert.sign:SM2|cert.sign:RSA|cert.sign:ECC|cert.sign:ML-DSA) return 0 ;;
    dual_cert.create:SM2) return 0 ;;
    crl.create:SM2|crl.create:RSA|crl.create:ECC|crl.create:ML-DSA) return 0 ;;
    *) return 1 ;;
  esac
}

# 生成私钥。所有 OpenSSL stderr 重定向至 /dev/null，避免污染 stderr JSON Lines。
pki_gen_key() {
  alg="$1"; out="$2"; in="$3"
  case "$alg" in
    SM2)
      "$OPENSSL_BIN" genpkey -algorithm SM2 -out "$out" 2>/dev/null || return 1
      ;;
    RSA)
      bits="$(json_get_def "$in" '.params.key_params.key_size' '2048')"
      case "$bits" in 2048|3072|4096) ;; *) return 2 ;; esac
      "$OPENSSL_BIN" genpkey -algorithm RSA -pkeyopt "rsa_keygen_bits:$bits" \
        -out "$out" 2>/dev/null || return 1
      ;;
    ECC)
      curve="$(json_get_def "$in" '.params.key_params.curve' 'prime256v1')"
      case "$curve" in prime256v1|P-256|secp384r1|P-384) ;; *) return 3 ;; esac
      "$OPENSSL_BIN" ecparam -name "$curve" -genkey -noout -out "$out" 2>/dev/null || return 1
      ;;
    ML-DSA)
      param="$(json_get_def "$in" '.params.key_params.parameter' 'ML-DSA-65')"
      "$OPENSSL_BIN" genpkey -algorithm "$param" -out "$out" 2>/dev/null || return 1
      ;;
    *)
      return 4
      ;;
  esac
}

# 生成 subject 字符串：/CN=xx/O=yy
pki_subject_arg() {
  in="$1"
  jq -r '.params.subject // {} | to_entries | map("\(.key)=\(.value)") | join("/") | "/" + .' "$in" \
    | sed 's#^//*#/#' \
    || printf '/CN=default\n'
}

pki_days() {
  in="$1"
  json_get_def "$in" '.params.validity_days' '365'
}

# 根据 CA 证书公钥算法选摘要参数。
# 原理：
#   1) 先提取 CA 证书公钥（SPKI），转成文本，判断是否为 SM2；
#   2) 若公钥判断不出，再从证书文本中查找 ASN1 OID / 曲线名；
#   3) 默认 SHA-256。
# 关键点：铜锁 x509 -text 输出中，SM2 证书的第一行 "Public Key Algorithm" 常为
#   id-ecPublicKey，而 SM2 出现在下一行 "ASN1 OID: SM2"，
#   因此不能只 grep 第一行。这里改用公钥文本判断，可靠且不依赖展示格式。
pki_digest_for_cert() {
  cert="$1"
  if [ ! -f "$cert" ]; then
    printf '%s\n' "-sha256"
    return 0
  fi

  # 优先：从公钥结构判断（最可靠）
  pub_text="$("$OPENSSL_BIN" x509 -in "$cert" -noout -pubkey 2>/dev/null \
              | "$OPENSSL_BIN" pkey -pubin -text -noout 2>/dev/null || true)"

  case "$pub_text" in
    *"SM2"*|*"sm2"*|*"sm2p256v1"*|*"1.2.156.10197.1.301"*)
      printf '%s\n' "-sm3"
      return 0
      ;;
  esac

  # 兜底：从证书文本里找 ASN1 OID / 曲线名 / 公钥算法
  cert_text="$("$OPENSSL_BIN" x509 -in "$cert" -noout -text 2>/dev/null || true)"
  case "$cert_text" in
    *"ASN1 OID: SM2"*|*"ASN1 OID: sm2"*|\
    *"Public Key Algorithm: SM2"*|*"Public Key Algorithm: sm2"*|\
    *"sm2p256v1"*|*"1.2.156.10197.1.301"*)
      printf '%s\n' "-sm3"
      return 0
      ;;
  esac

  # 默认：SHA-256（RSA/ECDSA）
  printf '%s\n' "-sha256"
}

pki_encrypt_key() {
  plain="$1"; key_ref="$2"
  enc="$CORE_ROOT/data/keys/${key_ref}.key.enc"
  pki_mkdir_for "$enc"
  key_store_encrypt "$plain" "$enc" >/dev/null
  printf '%s\n' "$enc"
}

pki_decrypt_key_to_tmp() {
  key_ref="$1"; tmp="$2"
  enc="$CORE_ROOT/data/keys/${key_ref}.key.enc"
  [ -f "$enc" ] || return 6
  key_store_decrypt "$enc" "$tmp" >/dev/null
  chmod 0600 "$tmp" 2>/dev/null || true
}

pki_cleanup_tmp() {
  if [ -n "${TMP_KEY:-}" ] && [ -f "${TMP_KEY:-}" ]; then
    shred -u "$TMP_KEY" 2>/dev/null || rm -f "$TMP_KEY"
  fi
  if [ -n "${TMP_PARENT_KEY:-}" ] && [ -f "${TMP_PARENT_KEY:-}" ]; then
    shred -u "$TMP_PARENT_KEY" 2>/dev/null || rm -f "$TMP_PARENT_KEY"
  fi
  if [ -n "${TMP_CA_KEY:-}" ] && [ -f "${TMP_CA_KEY:-}" ]; then
    shred -u "$TMP_CA_KEY" 2>/dev/null || rm -f "$TMP_CA_KEY"
  fi
  if [ -n "${TMP_SIGN_KEY:-}" ] && [ -f "${TMP_SIGN_KEY:-}" ]; then
    shred -u "$TMP_SIGN_KEY" 2>/dev/null || rm -f "$TMP_SIGN_KEY"
  fi
  if [ -n "${TMP_ENC_KEY:-}" ] && [ -f "${TMP_ENC_KEY:-}" ]; then
    shred -u "$TMP_ENC_KEY" 2>/dev/null || rm -f "$TMP_ENC_KEY"
  fi
  if [ -n "${TMP_DIR:-}" ] && [ -d "${TMP_DIR:-}" ]; then
    rm -rf "$TMP_DIR"
  fi
}

# 计算 params 的 SHA-256
pki_calc_params_digest() {
  in="$1"
  if [ ! -f "$in" ]; then
    printf 'sha256:unknown\n'
    return 0
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    h="$(jq -c '.params' "$in" 2>/dev/null | sha256sum | awk '{print $1}')"
  else
    h="$(jq -c '.params' "$in" 2>/dev/null | shasum -a 256 | awk '{print $1}')"
  fi
  [ -n "$h" ] || h="unknown"
  printf 'sha256:%s\n' "$h"
}

# 审计写入封装
pki_write_audit() {
  req="$1"; task="$2"; op="$3"; alg="$4"; result="$5"; dur="$6"; err="${7:-}"; in="${8:-}"
  pd="sha256:unknown"
  if [ -n "$in" ] && [ -f "$in" ]; then
    pd="$(pki_calc_params_digest "$in")"
  fi
  ACTOR_TYPE="${ACTOR_TYPE:-platform-backend}" ACTOR_ID="${ACTOR_ID:-}" \
    audit_write "$req" "$op" "$result" "$dur" "$err" "$task" "$alg" "$pd" "c07-20260921"
}

# 常用断言：文件存在
pki_require_file() {
  f="$1"; code="$2"; msg="$3"; out="$4"; exit_code="$5"
  if [ ! -f "$f" ]; then
    DETAIL="$(jq -n --arg p "$f" '{path:$p}')"
    pki_fail "$out" "$code" "$msg" "$DETAIL" "$exit_code"
  fi
}

# 常用断言：JSON 字段非空
pki_require_param() {
  in="$1"; path="$2"; field="$3"; out="$4"; exit_code="${5:-2}"
  v="$(json_get "$in" "$path")"
  if [ -z "$v" ] || [ "$v" = "null" ]; then
    DETAIL="$(jq -n --arg f "$field" '{field:$f}')"
    pki_fail "$out" "INVALID_PARAM" "missing required parameter: $field" "$DETAIL" "$exit_code"
  fi
}
