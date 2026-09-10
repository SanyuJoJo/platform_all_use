"""
认证模块 - 业务逻辑
基于 SQLAlchemy 异步会话，使用 auth_user / auth_role / auth_permission
/ auth_refresh_token 等表。
设计要点：
- 种子数据幂等且增量补充，不破坏已授权权限。
- Refresh Token 持久化到数据库，支持撤销与轮换。
- 刷新撤销使用条件 UPDATE，保证并发安全。
- 撤销批量操作使用批量 UPDATE。
- 关键认证事件通过结构化日志记录，后续可平滑接入日志审计模块。
v1.2 变更：
- P1-2：revoke_all_user_tokens 增加 `commit` 关键字参数（默认 True），
        允许调用方在同一事务内提交，避免用户管理模块删除用户时
        事务不原子。
"""
import hashlib
import logging
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional
from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from src.core.config import settings
from src.core.exceptions import PlatformException
from src.core.security import (
    create_access_token,
    decode_token,
    hash_password,
    verify_password,
)
from src.modules.auth.models import (
    Permission,
    RefreshToken,
    Role,
    User,
    UserRole,
)
logger = logging.getLogger(__name__)
ACCESS_TOKEN_EXPIRE_MINUTES = settings.ACCESS_TOKEN_EXPIRE_MINUTES
REFRESH_TOKEN_EXPIRE_DAYS = 7
# ---------------------------------------------------------------------------
# 日志辅助
# ---------------------------------------------------------------------------
def log_auth_event(
    action: str,
    *,
    user_id: Optional[int] = None,
    username: Optional[str] = None,
    status: str = "success",
    error_code: Optional[int] = None,
    ip: Optional[str] = None,
    user_agent: Optional[str] = None,
    detail: Optional[str] = None,
) -> None:
    """
    记录认证事件。
    当前使用 Python 结构化日志；待 audit_log 模块提供写入接口后，
    此函数将同时写入 audit_log_operation 表。
    """
    logger.info(
        "AUTH_EVENT action=%s user_id=%s username=%s status=%s "
        "error_code=%s ip=%s user_agent=%s detail=%s",
        action,
        user_id,
        username,
        status,
        error_code,
        ip,
        (user_agent or "")[:120],
        detail,
    )
# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
def _hash_token(token: str) -> str:
    """计算 Token 的 SHA-256 摘要，用于数据库存储。"""
    return hashlib.sha256(token.encode("utf-8")).hexdigest()
def _utcnow_naive() -> datetime:
    """返回 UTC naive datetime，与数据库 DateTime 字段保持一致。"""
    return datetime.now(timezone.utc).replace(tzinfo=None)
async def _load_user_with_rbac(db: AsyncSession, user_id: int) -> Optional[User]:
    result = await db.execute(
        select(User)
        .options(selectinload(User.roles).selectinload(Role.permissions))
        .where(User.id == user_id)
    )
    return result.scalar_one_or_none()
async def _load_user_by_username(db: AsyncSession, username: str) -> Optional[User]:
    result = await db.execute(
        select(User)
        .options(selectinload(User.roles).selectinload(Role.permissions))
        .where(User.username == username)
    )
    return result.scalar_one_or_none()
def serialize_user(user: User) -> Dict[str, Any]:
    """将 User ORM 对象序列化为认证模块统一用户字典。"""
    roles = [role.code for role in user.roles]
    permissions = sorted(
        {perm.code for role in user.roles for perm in role.permissions}
    )
    return {
        "id": user.id,
        "username": user.username,
        "nickname": user.nickname,
        "email": user.email,
        "avatar": user.avatar,
        "status": user.status,
        "roles": roles,
        "permissions": permissions,
    }
# ---------------------------------------------------------------------------
# 认证核心
# ---------------------------------------------------------------------------
async def authenticate_user(
    db: AsyncSession,
    username: str,
    password: str,
    ip: Optional[str] = None,
    user_agent: Optional[str] = None,
) -> Dict[str, Any]:
    """校验用户名密码，更新最后登录信息，返回用户信息字典。"""
    user = await _load_user_by_username(db, username)
    if not user or not verify_password(password, user.password_hash):
        log_auth_event(
            "login",
            username=username,
            status="fail",
            error_code=10001,
            ip=ip,
            user_agent=user_agent,
            detail="用户名或密码错误",
        )
        raise PlatformException(code=10001, message="用户名或密码错误", status_code=401)
    if user.status != 1:
        log_auth_event(
            "login",
            user_id=user.id,
            username=user.username,
            status="fail",
            error_code=10002,
            ip=ip,
            user_agent=user_agent,
            detail="用户已被禁用",
        )
        raise PlatformException(code=10002, message="用户已被禁用", status_code=403)
    user.last_login_at = _utcnow_naive()
    user.last_login_ip = ip
    await db.commit()
    user = await _load_user_with_rbac(db, user.id)
    user_info = serialize_user(user)
    log_auth_event(
        "login",
        user_id=user.id,
        username=user.username,
        status="success",
        ip=ip,
        user_agent=user_agent,
        detail="登录成功",
    )
    return user_info
async def create_tokens_for_user(
    db: AsyncSession,
    user_info: Dict[str, Any],
    ip: Optional[str] = None,
    user_agent: Optional[str] = None,
) -> Dict[str, Any]:
    """为用户签发 access_token 与 refresh_token，并持久化 Refresh Token。"""
    access_token = create_access_token(
        data={
            "sub": str(user_info["id"]),
            "username": user_info["username"],
            "roles": user_info["roles"],
            "permissions": user_info["permissions"],
            "type": "access",
        },
        expires_delta=timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES),
    )
    refresh_token = create_access_token(
        data={
            "sub": str(user_info["id"]),
            "type": "refresh",
        },
        expires_delta=timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS),
    )
    record = RefreshToken(
        user_id=user_info["id"],
        token_hash=_hash_token(refresh_token),
        expires_at=_utcnow_naive() + timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS),
        ip=ip,
        user_agent=(user_agent or "")[:255] or None,
    )
    db.add(record)
    await db.commit()
    return {
        "access_token": access_token,
        "refresh_token": refresh_token,
        "token_type": "bearer",
        "expires_in": ACCESS_TOKEN_EXPIRE_MINUTES,
    }
async def refresh_access_token(db: AsyncSession, refresh_token: str) -> Dict[str, Any]:
    """
    使用 refresh_token 换取新的 access_token。
    """
    try:
        payload = decode_token(refresh_token)
    except Exception:
        log_auth_event(
            "refresh", status="fail", error_code=10001, detail="Token 无法解析"
        )
        raise PlatformException(code=10001, message="Token无效或已过期", status_code=401)
    if payload.get("type") != "refresh":
        log_auth_event(
            "refresh", status="fail", error_code=10001, detail="Token 类型错误"
        )
        raise PlatformException(
            code=10001, message="无效的refresh token", status_code=401
        )
    user_id_raw = payload.get("sub")
    try:
        uid = int(user_id_raw)
    except (TypeError, ValueError):
        raise PlatformException(
            code=10001, message="无效的refresh token", status_code=401
        )
    token_hash = _hash_token(refresh_token)
    record = await db.scalar(
        select(RefreshToken).where(RefreshToken.token_hash == token_hash)
    )
    if not record or record.revoked:
        log_auth_event(
            "refresh", status="fail", error_code=10001, detail="Refresh Token 已失效"
        )
        raise PlatformException(
            code=10001, message="Refresh token 已失效", status_code=401
        )
    if record.user_id != uid:
        log_auth_event(
            "refresh",
            user_id=record.user_id,
            status="fail",
            error_code=10001,
            detail="Refresh Token 与用户不匹配",
        )
        raise PlatformException(
            code=10001, message="无效的refresh token", status_code=401
        )
    if record.expires_at < _utcnow_naive():
        log_auth_event(
            "refresh", status="fail", error_code=10001, detail="Refresh Token 已过期"
        )
        raise PlatformException(
            code=10001, message="Refresh token 已过期", status_code=401
        )
    user = await _load_user_with_rbac(db, uid)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if user.status != 1:
        raise PlatformException(code=10002, message="用户已被禁用", status_code=403)
    user_info = serialize_user(user)
    original_ip = record.ip
    original_user_agent = record.user_agent
    now = _utcnow_naive()
    revoke_result = await db.execute(
        update(RefreshToken)
        .where(
            RefreshToken.token_hash == token_hash,
            RefreshToken.revoked == 0,
        )
        .values(revoked=1, revoked_at=now)
    )
    if (revoke_result.rowcount or 0) == 0:
        await db.rollback()
        log_auth_event(
            "refresh",
            user_id=uid,
            status="fail",
            error_code=10001,
            detail="Refresh Token 已被并发撤销",
        )
        raise PlatformException(
            code=10001, message="Refresh token 已失效", status_code=401
        )
    access_token = create_access_token(
        data={
            "sub": str(user_info["id"]),
            "username": user_info["username"],
            "roles": user_info["roles"],
            "permissions": user_info["permissions"],
            "type": "access",
        },
        expires_delta=timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES),
    )
    new_refresh_token = create_access_token(
        data={
            "sub": str(user_info["id"]),
            "type": "refresh",
        },
        expires_delta=timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS),
    )
    new_record = RefreshToken(
        user_id=user_info["id"],
        token_hash=_hash_token(new_refresh_token),
        expires_at=now + timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS),
        ip=original_ip,
        user_agent=original_user_agent,
    )
    db.add(new_record)
    await db.commit()
    log_auth_event(
        "refresh",
        user_id=user_info["id"],
        username=user_info["username"],
        status="success",
    )
    return {
        "access_token": access_token,
        "refresh_token": new_refresh_token,
        "expires_in": ACCESS_TOKEN_EXPIRE_MINUTES,
    }
async def revoke_refresh_token(db: AsyncSession, refresh_token: str) -> None:
    """
    撤销指定的 Refresh Token（幂等）。
    """
    token_hash = _hash_token(refresh_token)
    now = _utcnow_naive()
    await db.execute(
        update(RefreshToken)
        .where(
            RefreshToken.token_hash == token_hash,
            RefreshToken.revoked == 0,
        )
        .values(revoked=1, revoked_at=now)
    )
    await db.commit()
async def revoke_all_user_tokens(
    db: AsyncSession,
    user_id: int,
    *,
    commit: bool = True,
) -> int:
    """
    撤销某用户的全部 Refresh Token，返回撤销数量。
    v1.2（P1-2）：
        新增 `commit` 关键字参数（默认 True），用于支持调用方在同一事务
        内完成撤销 + 其他操作。用户管理模块删除用户时传 `commit=False`，
        保证撤销与删除操作在同一事务提交，避免事务不原子。
    参数：
        db: 数据库会话
        user_id: 目标用户 ID
        commit: 是否立即提交；False 时由调用方负责提交或回滚
    返回：
        实际撤销的 Refresh Token 数量
    """
    now = _utcnow_naive()
    result = await db.execute(
        update(RefreshToken)
        .where(
            RefreshToken.user_id == user_id,
            RefreshToken.revoked == 0,
        )
        .values(revoked=1, revoked_at=now)
    )
    if commit:
        await db.commit()
    return result.rowcount or 0
async def get_user_info(db: AsyncSession, user_id: int) -> Dict[str, Any]:
    """获取用户完整信息。"""
    user = await _load_user_with_rbac(db, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    return serialize_user(user)
async def change_password(
    db: AsyncSession,
    user_id: int,
    old_password: str,
    new_password: str,
    confirm_password: str,
    ip: Optional[str] = None,
) -> None:
    """修改当前用户密码。成功后批量撤销该用户全部 Refresh Token。"""
    if new_password != confirm_password:
        log_auth_event(
            "change_password",
            user_id=user_id,
            status="fail",
            error_code=10004,
            ip=ip,
            detail="两次密码不一致",
        )
        raise PlatformException(code=10004, message="两次密码不一致", status_code=400)
    if not (6 <= len(new_password) <= 20):
        log_auth_event(
            "change_password",
            user_id=user_id,
            status="fail",
            error_code=10007,
            ip=ip,
            detail="新密码长度不符合要求",
        )
        raise PlatformException(code=10007, message="密码格式无效", status_code=400)
    user = await db.get(User, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if not verify_password(old_password, user.password_hash):
        log_auth_event(
            "change_password",
            user_id=user_id,
            username=user.username,
            status="fail",
            error_code=10003,
            ip=ip,
            detail="原密码错误",
        )
        raise PlatformException(code=10003, message="原密码错误", status_code=400)
    user.password_hash = hash_password(new_password)
    await db.commit()
    await revoke_all_user_tokens(db, user_id)
    log_auth_event(
        "change_password",
        user_id=user_id,
        username=user.username,
        status="success",
        ip=ip,
    )
# ---------------------------------------------------------------------------
# 种子数据
# ---------------------------------------------------------------------------
DEFAULT_PERMISSIONS: List[Dict[str, str]] = [
    {"code": "dashboard:view", "name": "查看仪表盘", "module_id": "platform", "resource": "dashboard", "action": "view"},
    {"code": "auth:user:view", "name": "查看用户", "module_id": "auth", "resource": "user", "action": "view"},
    {"code": "auth:user:create", "name": "创建用户", "module_id": "auth", "resource": "user", "action": "create"},
    {"code": "auth:user:edit", "name": "编辑用户", "module_id": "auth", "resource": "user", "action": "edit"},
    {"code": "auth:user:delete", "name": "删除用户", "module_id": "auth", "resource": "user", "action": "delete"},
    {"code": "auth:role:view", "name": "查看角色", "module_id": "auth", "resource": "role", "action": "view"},
    {"code": "auth:role:create", "name": "创建角色", "module_id": "auth", "resource": "role", "action": "create"},
    {"code": "auth:role:edit", "name": "编辑角色", "module_id": "auth", "resource": "role", "action": "edit"},
    {"code": "auth:role:delete", "name": "删除角色", "module_id": "auth", "resource": "role", "action": "delete"},
    {"code": "auth:permission:view", "name": "查看权限", "module_id": "auth", "resource": "permission", "action": "view"},
    {"code": "module_manager:module:view", "name": "查看模块", "module_id": "module_manager", "resource": "module", "action": "view"},
    {"code": "module_manager:module:create", "name": "安装模块", "module_id": "module_manager", "resource": "module", "action": "create"},
    {"code": "module_manager:module:edit", "name": "编辑模块", "module_id": "module_manager", "resource": "module", "action": "edit"},
    {"code": "module_manager:module:delete", "name": "卸载模块", "module_id": "module_manager", "resource": "module", "action": "delete"},
    {"code": "audit_log:log:view", "name": "查看日志", "module_id": "audit_log", "resource": "log", "action": "view"},
    {"code": "audit_log:log:export", "name": "导出日志", "module_id": "audit_log", "resource": "log", "action": "export"},
    {"code": "license:license:view", "name": "查看License", "module_id": "license", "resource": "license", "action": "view"},
    {"code": "license:license:create", "name": "管理License", "module_id": "license", "resource": "license", "action": "create"},
]
GUEST_PERMISSION_CODES = {
    "dashboard:view",
    "auth:user:view",
    "module_manager:module:view",
}
async def _upsert_permissions(
    db: AsyncSession,
) -> Dict[str, Permission]:
    """按 code upsert 核心权限。"""
    permissions_by_code: Dict[str, Permission] = {}
    for item in DEFAULT_PERMISSIONS:
        perm = await db.scalar(
            select(Permission).where(Permission.code == item["code"])
        )
        if not perm:
            perm = Permission(**item)
            db.add(perm)
            await db.flush()
        permissions_by_code[item["code"]] = perm
    return permissions_by_code
async def _get_role_with_permissions(
    db: AsyncSession, code: str
) -> Optional[Role]:
    """按 code 查询角色，预加载 permissions。"""
    return await db.scalar(
        select(Role)
        .options(selectinload(Role.permissions))
        .where(Role.code == code)
    )
async def _upsert_role(
    db: AsyncSession,
    code: str,
    name: str,
    description: str,
    initial_permissions: List[Permission],
) -> Role:
    """Upsert 系统角色。"""
    role = await _get_role_with_permissions(db, code)
    if not role:
        role = Role(
            name=name,
            code=code,
            description=description,
            is_system=1,
        )
        role.permissions = list(initial_permissions)
        db.add(role)
        await db.flush()
        return role
    existing_codes = {p.code for p in role.permissions}
    for perm in initial_permissions:
        if perm.code not in existing_codes:
            role.permissions.append(perm)
    return role
async def _upsert_user(
    db: AsyncSession,
    username: str,
    nickname: str,
    email: str,
) -> User:
    """Upsert 默认用户。"""
    user = await db.scalar(
        select(User).where(User.username == username)
    )
    if user:
        return user
    user = User(
        username=username,
        password_hash=hash_password("123456"),
        nickname=nickname,
        email=email,
        status=1,
    )
    db.add(user)
    await db.flush()
    return user
async def _ensure_user_role(
    db: AsyncSession,
    user_id: int,
    role_id: int,
) -> None:
    """幂等地确保 user_id 与 role_id 存在关联。"""
    exists = await db.scalar(
        select(UserRole).where(
            UserRole.user_id == user_id,
            UserRole.role_id == role_id,
        )
    )
    if not exists:
        db.add(UserRole(user_id=user_id, role_id=role_id))
async def ensure_auth_seed_data(db: AsyncSession) -> None:
    """初始化认证模块种子数据（幂等、增量）。"""
    permissions_by_code = await _upsert_permissions(db)
    admin_role = await _upsert_role(
        db,
        code="admin",
        name="管理员",
        description="系统内置管理员，拥有全部权限",
        initial_permissions=list(permissions_by_code.values()),
    )
    guest_initial = [
        p for code, p in permissions_by_code.items() if code in GUEST_PERMISSION_CODES
    ]
    guest_role = await _upsert_role(
        db,
        code="guest",
        name="访客",
        description="系统内置访客，只读权限",
        initial_permissions=guest_initial,
    )
    admin_user = await _upsert_user(
        db,
        username="admin",
        nickname="管理员",
        email="admin@example.com",
    )
    guest_user = await _upsert_user(
        db,
        username="guest",
        nickname="访客",
        email="guest@example.com",
    )
    await _ensure_user_role(db, admin_user.id, admin_role.id)
    await _ensure_user_role(db, guest_user.id, guest_role.id)
    await db.commit()
    logger.info("认证种子数据初始化完成（幂等、增量）")
