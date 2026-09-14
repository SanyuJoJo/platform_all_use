"""License 管理模块 - 业务逻辑。
v1.1 修复（评审报告闭环）：
- P0-1：删除 _persist_license 中无意义查询（占位 select）。
- P0-2：捕获 IntegrityError，映射为 50004。
- P0-3：activate_license 前置校验 machine_code 与当前一致；强制使用当前机器码。
- P0-4：_persist_license 事务顺序调整：先 UPDATE 旧 active，再 add 新记录；
        显式同步 updated_at。
- P0-5：补齐 max_users / authorized_modules / license_key / 时间逻辑校验。
- P1-1：删除 check_module_authorized 冗余 is_active 判断。
- P1-2：收窄 get_module_authorization 异常捕获（仅 ImportError）。
- P1-3：_count_current_users 统计全部用户（含禁用）。
- P1-4：import_license 空字节健壮判断。
- P1-7：get_license_status 移除 machine_code / current_machine_code 响应字段。
- P1-9：from sqlalchemy import update 移至文件顶部。
- P1-10：existing 仅 is_active==1 时抛 50004；inactive 记录允许重新激活。
- P2-1：is_expiring_soon 条件改为 0 <= days_remaining <= EXPIRING_SOON_DAYS。
- P2-8：外部服务返回 data 结构校验。
v1.2 兼容性说明（R-3）：
- 数据库层新增部分唯一索引 uq_license_license_active。
- 业务层无需修改：并发冲突时由 IntegrityError 统一映射为 50004。
"""
import json
import logging
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import httpx
from sqlalchemy import func, select, update
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.config import settings
from src.core.exceptions import PlatformException
from src.modules.auth.models import User
from src.modules.auth.service import log_auth_event
from src.modules.license.constants import (
    CORE_MODULES_BYPASS,
    EXPIRING_SOON_DAYS,
    LICENSE_KEY_MAX_LEN,
    LICENSE_KEY_MIN_LEN,
    LICENSE_TYPES,
    REQUIRED_PAYLOAD_FIELDS,
)
from src.modules.license.models import License
from src.modules.license.validator import (
    get_machine_code,
    parse_license_content,
    verify_signature,
)

logger = logging.getLogger(__name__)
# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
def _utcnow_naive() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)
def _parse_iso_datetime(value: Any, field: str) -> datetime:
    """解析 ISO 8601 字符串为 naive datetime。"""
    if isinstance(value, datetime):
        dt = value
    elif isinstance(value, str):
        try:
            normalized = (
                value.replace("Z", "+00:00") if value.endswith("Z") else value
            )
            dt = datetime.fromisoformat(normalized)
        except ValueError as exc:
            raise PlatformException(
                code=50001,
                message=f"License 字段 {field} 时间格式无效：{value}",
                status_code=400,
            ) from exc
    else:
        raise PlatformException(
            code=50001,
            message=f"License 字段 {field} 必须为时间字符串",
            status_code=400,
        )
    if dt.tzinfo is not None:
        dt = dt.astimezone(timezone.utc).replace(tzinfo=None)
    return dt
async def _load_active_license(db: AsyncSession) -> Optional[License]:
    """查询当前生效的 License（is_active=1，取 expires_at 最新一条）。"""
    result = await db.execute(
        select(License)
        .where(License.is_active == 1)
        .order_by(License.expires_at.desc())
        .limit(1)
    )
    return result.scalar_one_or_none()
async def _count_current_users(db: AsyncSession) -> int:
    """
    统计当前用户数。
    v1.1（P1-3）：统计全部用户（含禁用），与产品对"用户数"的定义对齐。
    License 限制的是创建的用户总数，与用户状态无关。
    """
    result = await db.execute(select(func.count(User.id)))
    return int(result.scalar_one() or 0)
def _compute_days_remaining(expires_at: datetime) -> int:
    """计算剩余天数，负数按 0 处理。"""
    delta = expires_at - _utcnow_naive()
    days = delta.days
    return max(days, 0)
def _is_license_valid(
    license_obj: License, current_machine_code: str
) -> bool:
    """
    判断 License 是否有效：
    - is_active == 1
    - expires_at > now()
    - 若绑定了 machine_code，需与当前机器码一致
    """
    if license_obj.is_active != 1:
        return False
    if license_obj.expires_at <= _utcnow_naive():
        return False
    if license_obj.machine_code and license_obj.machine_code != current_machine_code:
        return False
    return True
# ---------------------------------------------------------------------------
# 状态查询（v1.1 P1-7：移除 machine_code / current_machine_code）
# ---------------------------------------------------------------------------
async def get_license_status(db: AsyncSession) -> Dict[str, Any]:
    """
    获取当前 License 状态。
    若未导入或已失效（is_active=0），返回 50009。
    过期 License 仍返回状态，但 is_expired=True / is_valid=False。
    v1.1（P1-7）：响应不再包含 machine_code / current_machine_code。
    """
    license_obj = await _load_active_license(db)
    if not license_obj:
        raise PlatformException(
            code=50009, message="License不存在", status_code=404
        )
    current_machine_code = get_machine_code()
    current_users = await _count_current_users(db)
    days_remaining = _compute_days_remaining(license_obj.expires_at)
    is_expired = license_obj.expires_at <= _utcnow_naive()
    is_valid = _is_license_valid(license_obj, current_machine_code)
    # v1.1（P2-1）：0 天也视为即将过期
    is_expiring_soon = (
        not is_expired
        and 0 <= days_remaining <= EXPIRING_SOON_DAYS
    )
    # 机器码仅写入服务端日志，便于排障
    logger.debug(
        "License 状态：key=%s bound_machine=%s current_machine=%s",
        license_obj.license_key,
        license_obj.machine_code,
        current_machine_code,
    )
    return {
        "is_valid": is_valid,
        "license_type": license_obj.license_type,
        "expires_at": license_obj.expires_at,
        "days_remaining": days_remaining,
        "max_users": license_obj.max_users,
        "current_users": current_users,
        "authorized_modules": list(license_obj.authorized_modules or []),
        "is_expired": is_expired,
        "is_expiring_soon": is_expiring_soon,
    }
# ---------------------------------------------------------------------------
# 数据校验（v1.1 P0-5 新增）
# ---------------------------------------------------------------------------
def _validate_license_key(payload: Dict[str, Any]) -> str:
    """校验 license_key（v1.1 P0-5）。"""
    license_key = payload.get("license_key")
    if not isinstance(license_key, str):
        raise PlatformException(
            code=50001,
            message="License 字段 license_key 必须为字符串",
            status_code=400,
        )
    if not (LICENSE_KEY_MIN_LEN <= len(license_key) <= LICENSE_KEY_MAX_LEN):
        raise PlatformException(
            code=50001,
            message=(
                f"License 字段 license_key 长度必须为 "
                f"{LICENSE_KEY_MIN_LEN}-{LICENSE_KEY_MAX_LEN} 位"
            ),
            status_code=400,
        )
    return license_key
def _validate_max_users(payload: Dict[str, Any]) -> Optional[int]:
    """校验 max_users（v1.1 P0-5，排除 bool）。"""
    max_users = payload.get("max_users")
    if max_users is None:
        return None
    # bool 是 int 的子类，需显式排除
    if isinstance(max_users, bool) or not isinstance(max_users, int):
        raise PlatformException(
            code=50001,
            message="License 字段 max_users 必须为正整数或 null",
            status_code=400,
        )
    if max_users <= 0:
        raise PlatformException(
            code=50001,
            message="License 字段 max_users 必须为正整数或 null",
            status_code=400,
        )
    return max_users
def _validate_authorized_modules(payload: Dict[str, Any]) -> List[str]:
    """校验 authorized_modules（v1.1 P0-5）。"""
    authorized_modules = payload.get("authorized_modules") or []
    if not isinstance(authorized_modules, list):
        raise PlatformException(
            code=50001,
            message="License 字段 authorized_modules 必须为数组",
            status_code=400,
        )
    for m in authorized_modules:
        if not isinstance(m, str) or not m:
            raise PlatformException(
                code=50001,
                message="License 字段 authorized_modules 元素必须为非空字符串",
                status_code=400,
            )
    return list(authorized_modules)
def _validate_machine_code_field(payload: Dict[str, Any]) -> Optional[str]:
    """校验 machine_code 字段类型（v1.1 P0-5）。"""
    bound_machine = payload.get("machine_code")
    if bound_machine is None:
        return None
    if not isinstance(bound_machine, str):
        raise PlatformException(
            code=50001,
            message="License 字段 machine_code 必须为字符串或 null",
            status_code=400,
        )
    return bound_machine
# ---------------------------------------------------------------------------
# 导入 / 激活
# ---------------------------------------------------------------------------
async def _persist_license(
    db: AsyncSession,
    payload: Dict[str, Any],
    operator: Optional[Dict[str, Any]] = None,
) -> License:
    """
    将已校验的 payload 落库。
    v1.1（P0-4）：事务顺序调整
        1. 校验参数
        2. 查询 existing
        3. 先 UPDATE 旧 active（除本记录外）→ is_active=0, updated_at=now
        4. 再 add 新记录 / 更新 existing
        5. commit（捕获 IntegrityError → 50004）
    v1.1（P0-5）：补齐校验
    v1.1（P1-10）：existing.is_active==1 才抛 50004
    v1.2（R-3）：依赖数据库部分唯一索引 uq_license_license_active
        并发导入不同 license_key 时，第二个事务的 INSERT 会因
        部分唯一索引冲突抛出 IntegrityError，被统一映射为 50004。
    """
    # ---- 1. license_type 白名单 ----
    license_type = payload.get("license_type")
    if license_type not in LICENSE_TYPES:
        raise PlatformException(
            code=50010,
            message=f"License 类型不支持：{license_type}",
            status_code=400,
        )
    # ---- 2. 必需字段 ----
    for field in REQUIRED_PAYLOAD_FIELDS:
        if field not in payload or payload[field] in (None, ""):
            raise PlatformException(
                code=50001,
                message=f"License 缺少必需字段：{field}",
                status_code=400,
            )
    # ---- 3. license_key 校验 ----
    license_key = _validate_license_key(payload)
    # ---- 4. 时间解析与逻辑校验 ----
    issued_at = _parse_iso_datetime(payload["issued_at"], "issued_at")
    expires_at = _parse_iso_datetime(payload["expires_at"], "expires_at")
    if issued_at >= expires_at:
        raise PlatformException(
            code=50001,
            message="License 字段 issued_at 必须早于 expires_at",
            status_code=400,
        )
    if expires_at <= _utcnow_naive():
        raise PlatformException(
            code=50002, message="License 已过期", status_code=400
        )
    # ---- 5. machine_code 校验 ----
    current_machine_code = get_machine_code()
    bound_machine = _validate_machine_code_field(payload)
    if bound_machine and bound_machine != current_machine_code:
        raise PlatformException(
            code=50006,
            message=(
                f"License 绑定机器码不匹配"
                f"（license={bound_machine}, current={current_machine_code}）"
            ),
            status_code=400,
        )
    # ---- 6. max_users 校验 ----
    max_users = _validate_max_users(payload)
    # ---- 7. authorized_modules 校验 ----
    authorized_modules = _validate_authorized_modules(payload)
    # ---- 8. 查询 existing ----
    existing = await db.scalar(
        select(License).where(License.license_key == license_key)
    )
    now = _utcnow_naive()
    # ---- 9. existing 且 is_active==1 → 50004（P1-10） ----
    if existing is not None and existing.is_active == 1:
        raise PlatformException(
            code=50004, message="License 已激活", status_code=409
        )
    # ---- 10. 先 UPDATE 旧 active（除本记录外）为 inactive ----
    # v1.1（P0-4）：顺序调整，不依赖 autoflush；显式同步 updated_at
    await db.execute(
        update(License)
        .where(License.is_active == 1, License.license_key != license_key)
        .values(is_active=0, updated_at=now)
    )
    # ---- 11. 新增 / 重新激活 ----
    try:
        if existing is not None:
            # v1.1（P1-10）：inactive 记录允许重新激活
            existing.license_type = license_type
            existing.max_users = max_users
            existing.authorized_modules = list(authorized_modules)
            existing.machine_code = bound_machine
            existing.issued_at = issued_at
            existing.expires_at = expires_at
            existing.is_active = 1
            existing.activated_at = now
            existing.updated_at = now
            license_obj = existing
        else:
            license_obj = License(
                license_key=license_key,
                license_type=license_type,
                max_users=max_users,
                authorized_modules=list(authorized_modules),
                machine_code=bound_machine,
                issued_at=issued_at,
                expires_at=expires_at,
                is_active=1,
                activated_at=now,
            )
            db.add(license_obj)
        # v1.1（P0-2）：捕获 IntegrityError（并发导入相同 license_key）
        # v1.2（R-3）：部分唯一索引冲突也在此捕获
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise PlatformException(
            code=50004, message="License 已激活", status_code=409
        ) from exc
    await db.refresh(license_obj)
    if operator:
        log_auth_event(
            "license_import",
            user_id=operator["id"],
            username=operator["username"],
            status="success",
            detail=(
                f"导入 License {license_key} "
                f"（type={license_type}, "
                f"modules={len(authorized_modules)}, "
                f"expires={expires_at.isoformat()}）"
            ),
        )
    logger.info(
        "License 导入成功：key=%s type=%s expires=%s",
        license_key,
        license_type,
        expires_at,
    )
    return license_obj
async def import_license(
    db: AsyncSession,
    *,
    file_content: Optional[bytes],
    activation_code: Optional[str],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """
    导入 License。
    两种方式（二选一）：
    - file_content：License 文件内容（.lic / .json）
    - activation_code：在线激活码
    v1.1（P1-4）：空字节健壮判断。
    """
    has_file = file_content is not None and len(file_content) > 0
    has_code = bool(activation_code and activation_code.strip())
    if not has_file and not has_code:
        raise PlatformException(
            code=50001,
            message="必须提供 license_file 或 activation_code",
            status_code=400,
        )
    if has_file:
        parsed = parse_license_content(file_content)
    else:
        file_content = await _fetch_license_by_activation_code(
            activation_code=activation_code or "",
            machine_code=get_machine_code(),
        )
        parsed = parse_license_content(file_content)
    payload = parsed["payload"]
    signature = parsed["signature"]
    if not verify_signature(payload, signature, settings.LICENSE_SECRET_KEY):
        raise PlatformException(
            code=50003, message="License 签名验证失败", status_code=400
        )
    license_obj = await _persist_license(db, payload, operator)
    return {
        "id": license_obj.id,
        "license_key": license_obj.license_key,
        "license_type": license_obj.license_type,
        "expires_at": license_obj.expires_at,
        "authorized_modules": list(license_obj.authorized_modules or []),
        "max_users": license_obj.max_users,
        "machine_code": license_obj.machine_code,
    }
async def activate_license(
    db: AsyncSession,
    *,
    activation_code: str,
    machine_code: str,
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """
    在线激活。
    v1.1（P0-3）：前置校验请求 machine_code 与当前机器码一致；
        强制使用当前机器码请求外部服务（防篡改）。
    """
    current_machine_code = get_machine_code()
    # ---- P0-3：前置校验 ----
    if machine_code != current_machine_code:
        raise PlatformException(
            code=50006,
            message=(
                f"机器码不匹配（请求={machine_code}, "
                f"当前={current_machine_code}）"
            ),
            status_code=400,
        )
    # ---- 强制使用当前机器码请求外部服务 ----
    file_content = await _fetch_license_by_activation_code(
        activation_code=activation_code,
        machine_code=current_machine_code,
    )
    parsed = parse_license_content(file_content)
    payload = parsed["payload"]
    signature = parsed["signature"]
    if not verify_signature(payload, signature, settings.LICENSE_SECRET_KEY):
        raise PlatformException(
            code=50003, message="License 签名验证失败", status_code=400
        )
    license_obj = await _persist_license(db, payload, operator)
    return {
        "id": license_obj.id,
        "license_key": license_obj.license_key,
        "license_type": license_obj.license_type,
        "expires_at": license_obj.expires_at,
        "authorized_modules": list(license_obj.authorized_modules or []),
        "max_users": license_obj.max_users,
        "machine_code": license_obj.machine_code,
    }
async def _fetch_license_by_activation_code(
    *, activation_code: str, machine_code: str
) -> bytes:
    """
    调用外部激活服务获取 License 文件内容。
    v1.1（P2-8）：校验响应 data 必须包含 payload 和 signature。
    """
    url = (settings.LICENSE_ACTIVATION_URL or "").strip()
    if not url:
        raise PlatformException(
            code=50005,
            message="激活码无效（未配置在线激活服务）",
            status_code=400,
        )
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.post(
                f"{url.rstrip('/')}/activate",
                json={
                    "activation_code": activation_code,
                    "machine_code": machine_code,
                },
            )
    except httpx.HTTPError as exc:
        raise PlatformException(
            code=50005,
            message=f"在线激活服务不可用：{exc}",
            status_code=400,
        ) from exc
    if resp.status_code != 200:
        raise PlatformException(
            code=50005,
            message=f"在线激活失败（HTTP {resp.status_code}）",
            status_code=400,
        )
    try:
        body = resp.json()
    except Exception as exc:
        raise PlatformException(
            code=50005,
            message=f"在线激活响应不是合法 JSON：{exc}",
            status_code=400,
        ) from exc
    if not isinstance(body, dict) or body.get("code") != 0:
        msg = body.get("message") if isinstance(body, dict) else "未知错误"
        raise PlatformException(
            code=50005,
            message=f"激活码无效：{msg}",
            status_code=400,
        )
    data = body.get("data")
    if not isinstance(data, dict):
        raise PlatformException(
            code=50005,
            message="在线激活响应缺少 data",
            status_code=400,
        )
    # v1.1（P2-8）：结构校验
    if not isinstance(data.get("payload"), dict) or not isinstance(
        data.get("signature"), str
    ):
        raise PlatformException(
            code=50005,
            message="在线激活响应缺少 payload 或 signature",
            status_code=400,
        )
    return json.dumps(data, ensure_ascii=False).encode("utf-8")
# ---------------------------------------------------------------------------
# 模块授权
# ---------------------------------------------------------------------------
async def get_module_authorization(db: AsyncSession) -> List[Dict[str, Any]]:
    """
    获取所有已安装模块的授权状态。
    v1.1（P1-2）：收窄异常捕获——仅 ImportError 时降级；
        其他异常记录日志后向上抛出，避免掩盖真实错误。
    """
    license_obj = await _load_active_license(db)
    authorized_modules: set[str] = set()
    expires_at: Optional[datetime] = None
    if license_obj:
        authorized_modules = set(license_obj.authorized_modules or [])
        expires_at = license_obj.expires_at
    # 尝试导入模块管理模块的 Module 模型（仅 ImportError 时降级）
    try:
        from src.modules.module_manager.models import Module
    except ImportError as exc:
        logger.warning(
            "模块管理模块不可用，降级为 License 授权列表本身：%s", exc
        )
        return [
            {
                "module_id": mid,
                "module_name": mid,
                "is_authorized": True,
                "expires_at": expires_at,
            }
            for mid in sorted(authorized_modules)
        ]
    result = await db.execute(select(Module).order_by(Module.id.asc()))
    modules = list(result.scalars().all())
    return [
        {
            "module_id": m.id,
            "module_name": m.name,
            "is_authorized": m.id in authorized_modules
            or m.id in CORE_MODULES_BYPASS,
            "expires_at": expires_at,
        }
        for m in modules
    ]
async def check_module_authorized(
    db: AsyncSession, module_id: str
) -> None:
    """
    校验模块是否已授权。
    v1.1（P1-1）：删除冗余 is_active != 1 判断
        （_load_active_license 已过滤 is_active==1）。
    """
    if module_id in CORE_MODULES_BYPASS:
        return
    license_obj = await _load_active_license(db)
    if not license_obj:
        raise PlatformException(
            code=50009, message="License不存在", status_code=404
        )
    if license_obj.expires_at <= _utcnow_naive():
        raise PlatformException(
            code=50002, message="License 已过期", status_code=400
        )
    if license_obj.machine_code and license_obj.machine_code != get_machine_code():
        raise PlatformException(
            code=50006, message="机器码不匹配", status_code=400
        )
    authorized = set(license_obj.authorized_modules or [])
    if module_id not in authorized:
        raise PlatformException(
            code=50007,
            message=f"模块未授权：{module_id}",
            status_code=403,
        )
# ---------------------------------------------------------------------------
# 用户配额
# ---------------------------------------------------------------------------
async def check_user_quota(db: AsyncSession) -> None:
    """
    校验当前用户数是否超过 License 限制。
    v1.1（P1-3）：统计全部用户（含禁用）。
    """
    license_obj = await _load_active_license(db)
    if not license_obj:
        raise PlatformException(
            code=50009, message="License不存在", status_code=404
        )
    if license_obj.expires_at <= _utcnow_naive():
        raise PlatformException(
            code=50002, message="License 已过期", status_code=400
        )
    if license_obj.max_users is None:
        return
    current_users = await _count_current_users(db)
    if current_users >= license_obj.max_users:
        raise PlatformException(
            code=50008,
            message=(
                f"用户数超限：当前 {current_users}，"
                f"License 上限 {license_obj.max_users}"
            ),
            status_code=400,
        )
# ---------------------------------------------------------------------------
# 启动校验
# ---------------------------------------------------------------------------
async def verify_license_on_startup(db: AsyncSession) -> Dict[str, Any]:
    """
    启动时校验 License 状态。
    - 不阻断启动（允许运维通过 API 导入/更新 License）。
    - 返回状态字典，供调用方记录日志或决定是否阻断。
    """
    try:
        status = await get_license_status(db)
        if status["is_expired"]:
            logger.error(
                "启动校验：License 已过期（expires_at=%s）",
                status["expires_at"],
            )
        elif not status["is_valid"]:
            logger.error(
                "启动校验：License 无效（machine_code 不匹配或 is_active=0）"
            )
        elif status["is_expiring_soon"]:
            logger.warning(
                "启动校验：License 即将过期（剩余 %s 天）",
                status["days_remaining"],
            )
        else:
            logger.info(
                "启动校验：License 有效（type=%s, 剩余 %s 天）",
                status["license_type"],
                status["days_remaining"],
            )
        return {"ok": True, "status": status}
    except PlatformException as exc:
        if exc.code == 50009:
            logger.warning("启动校验：未导入 License（code=50009）")
        else:
            logger.error("启动校验失败：code=%s message=%s", exc.code, exc.message)
        return {"ok": False, "code": exc.code, "message": exc.message}
    except Exception as exc:
        logger.error("启动校验异常：%s", exc, exc_info=True)
        return {"ok": False, "code": 90000, "message": str(exc)}
