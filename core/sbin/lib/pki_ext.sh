#!/bin/sh
# core/sbin/lib/pki_ext.sh
# P1-P2 公共扩展函数：算法校验、路径校验、口令文件校验、PQC 密钥生成。
# 依赖 pki_common.sh 提供的 pki_fail / pki_gen_key / json_get / json_get_def。

_CORE_PKI_EXT_LOADED="${_CORE_PKI_EXT_LOADED:-0}"
if [ "$_CORE_PKI_EXT_LOADED" = "1" ]; then
  return 0 2>/dev/null || true
fi
_CORE_PKI_EXT_LOADED=1

# -----------------------------------------------------------------------------
# 算法白名单校验：按 operation_id + algorithm 匹配
# -----------------------------------------------------------------------------
pki_ext_check_algorithm() {
  op="$1"; alg="$2"
  case "$op:$alg" in
    cert.convert:*|cert.parse:*) return 0 ;;
    key.manage:SM2|key.manage:RSA|key.manage:ECC|key.manage:ML-KEM|key.manage:ML-DSA|key.manage:SLH-DSA) return 0 ;;
    pqc.cert.create:ML-KEM|pqc.cert.create:ML-DSA|pqc.cert.create:SLH-DSA) return 0 ;;
    chain.verify:*|batch.execute:*|ssl.config.generate:*) return 0 ;;
    crypto.service:SM2|crypto.service:RSA|crypto.service:ECC|crypto.service:ML-DSA|crypto.service:SLH-DSA) return 0 ;;
    *) return 1 ;;
  esac
}

# -----------------------------------------------------------------------------
# 路径归一化：相对路径拼到 $CORE_ROOT 下
# -----------------------------------------------------------------------------
pki_ext_abs_path() {
  p="$1"
  case "$p" in
    /*) printf '%s\n' "$p" ;;
    *)  printf '%s\n' "$CORE_ROOT/$p" ;;
  esac
}

# -----------------------------------------------------------------------------
# 路径必须在 core/data 或 core/tmp 下
# 成功后 stdout 输出绝对路径
# -----------------------------------------------------------------------------
pki_ext_require_under_core() {
  path="$1"; field="$2"; out="$3"
  abs="$(pki_ext_abs_path "$path")"
  case "$abs" in
    "$CORE_ROOT/data/"*|"$CORE_ROOT/tmp/"*)
      printf '%s\n' "$abs"
      ;;
    *)
      DETAIL="$(jq -n --arg f "$field" --arg p "$path" '{field:$f,path:$p}')"
      pki_fail "$out" "PATH_NOT_ALLOWED" "path not allowed" "$DETAIL" 4
      ;;
  esac
}

# -----------------------------------------------------------------------------
# 口令文件校验：
#   1) 先归一化为绝对路径（关键修复：相对路径也能通过前缀匹配）
#   2) 必须在 core/tmp 或 core/data 下
#   3) 权限必须为 0600
# -----------------------------------------------------------------------------
pki_ext_require_password_file() {
  f="$1"; out="$2"

  abs="$(pki_ext_abs_path "$f")"

  [ -f "$abs" ] || {
    DETAIL="$(jq -n --arg f "$f" '{field:"password_file",reason:"not_found"}')"
    pki_fail "$out" "INVALID_PARAM" "password_file not found" "$DETAIL" 2
  }

  case "$abs" in
    "$CORE_ROOT/tmp/"*|"$CORE_ROOT/data/"*) ;;
    *)
      DETAIL="$(jq -n --arg f "$f" '{field:"password_file",reason:"path_not_allowed"}')"
      pki_fail "$out" "PATH_NOT_ALLOWED" "password_file path not allowed" "$DETAIL" 4
      ;;
  esac

  mode="$(stat -c '%a' "$abs" 2>/dev/null || stat -f '%Lp' "$abs" 2>/dev/null || echo '?')"
  [ "$mode" = "600" ] || {
    DETAIL="$(jq -n --arg m "$mode" '{field:"password_file",mode:$m}')"
    pki_fail "$out" "INVALID_PARAM" "password_file permission must be 600" "$DETAIL" 2
  }

  printf '%s\n' "$abs"
}

# -----------------------------------------------------------------------------
# PQC 密钥生成：在 pki_gen_key 基础上扩展 ML-KEM / SLH-DSA
# -----------------------------------------------------------------------------
pki_ext_gen_key() {
  alg="$1"; out="$2"; in="$3"
  case "$alg" in
    ML-KEM)
      param="$(json_get_def "$in" '.params.key_params.parameter' 'ML-KEM-768')"
      "$OPENSSL_BIN" genpkey -algorithm "$param" -out "$out" 2>/dev/null || return 1
      ;;
    SLH-DSA)
      param="$(json_get_def "$in" '.params.key_params.parameter' 'SLH-DSA-SHA2-128s')"
      "$OPENSSL_BIN" genpkey -algorithm "$param" -out "$out" 2>/dev/null || return 1
      ;;
    *)
      pki_gen_key "$alg" "$out" "$in"
      ;;
  esac
}
