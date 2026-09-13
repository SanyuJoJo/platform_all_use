"""License 管理模块 - 签名验证与机器码生成。
设计要点：
- 签名算法：HMAC-SHA256（初版），预留 RSA 升级点。
- 规范化 JSON：json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
- 使用 hmac.compare_digest 防止时序攻击。
- 机器码：MAC 地址 + 平台信息 → SHA-256 → 前 32 位大写 hex。
"""
import hashlib
import hmac
import json
import logging
import platform
import uuid
from typing import Any, Dict
from src.core.config import settings
from src.core.exceptions import PlatformException
logger = logging.getLogger(__name__)
def _canonical_json(payload: Dict[str, Any]) -> str:
    """规范化 JSON，保证签名稳定性。"""
    return json.dumps(
        payload,
        sort_keys=True,
        separators=(",", ":"),
        ensure_ascii=False,
    )
def compute_signature(payload: Dict[str, Any], secret: str) -> str:
    """
    计算 payload 的 HMAC-SHA256 签名。
    参数：
        payload: License payload 字典
        secret: 签名密钥（LICENSE_SECRET_KEY）
    返回：
        十六进制签名字符串
    """
    canonical = _canonical_json(payload)
    return hmac.new(
        secret.encode("utf-8"),
        canonical.encode("utf-8"),
        hashlib.sha256,
    ).hexdigest()
def verify_signature(
    payload: Dict[str, Any], signature: str, secret: str
) -> bool:
    """
    验证 payload 的签名。
    使用 hmac.compare_digest 防止时序攻击。
    """
    expected = compute_signature(payload, secret)
    return hmac.compare_digest(expected, signature)
def get_machine_code() -> str:
    """
    生成当前机器码。
    - 支持 LICENSE_MACHINE_CODE_OVERRIDE 环境变量覆盖（开发/测试）。
    - 否则基于 MAC 地址 + 平台信息生成，SHA-256 后取前 32 位大写 hex。
    """
    override = (settings.LICENSE_MACHINE_CODE_OVERRIDE or "").strip()
    if override:
        return override
    mac = uuid.getnode()
    parts = [
        f"{mac:012x}",
        platform.system(),
        platform.machine(),
    ]
    raw = "|".join(parts).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()[:32].upper()
def parse_license_content(content: bytes) -> Dict[str, Any]:
    """
    解析 License 文件内容，返回 {"payload": {...}, "signature": "..."}。
    失败时抛出 50001。
    """
    try:
        text = content.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise PlatformException(
            code=50001,
            message="License 文件编码无效（需 UTF-8）",
            status_code=400,
        ) from exc
    try:
        data = json.loads(text)
    except json.JSONDecodeError as exc:
        raise PlatformException(
            code=50001,
            message=f"License 文件不是合法 JSON：{exc}",
            status_code=400,
        ) from exc
    if not isinstance(data, dict):
        raise PlatformException(
            code=50001, message="License 文件结构无效", status_code=400
        )
    payload = data.get("payload")
    signature = data.get("signature")
    if not isinstance(payload, dict):
        raise PlatformException(
            code=50001, message="License 文件缺少 payload", status_code=400
        )
    if not isinstance(signature, str) or not signature:
        raise PlatformException(
            code=50001, message="License 文件缺少 signature", status_code=400
        )
    return {"payload": payload, "signature": signature}
