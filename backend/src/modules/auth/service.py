"""
认证模块 - 业务逻辑

基于 SQLAlchemy 异步会话，使用 auth_user / auth_role / auth_permission
/ auth_refresh_token 等表。

设计要点：
- 种子数据幂等且增量补充，不破坏已授权权限。
- Refresh Token 持久化到数据库，支持撤销与轮换（每次刷新更换新 Token）。
- 刷新撤销使用条件 UPDATE，保证并发安全（N-2）。
- 撤销批量操作使用批量 UPDATE（N-6）。
- 关键认证事件通过结构化日志记录，后续可平滑接入日志审计模块。

v1.3.1 修复：
- 修复 SQLAlchemy 2.0 异步模式下种子数据初始化的 MissingGreenlet 错误：
  1. `_upsert_role` 中新建角色的 secondary 关系赋值改为在 `db.add()` 之前；
  2. `ensure_auth_seed_data` 中用户-角色关联判断改为查询 `auth_user_role`
     中间表，避免访问未预加载的 `user.roles` 关系。

v1.3.2 修复：
- 修复 `refresh_access_token` 中 UPDATE 语句后访问过期 ORM 对象属性
  导致的 MissingGreenlet → HTTP 500：在 UPDATE 之前提取所有需要的
  普通 Python 数据（user_info / original_ip / original_user_agent）。
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

    v1.2 起：
    - 校验 record.user_id 与 Token sub 一致（纵深防御）；
    - 旧 Refresh Token 立即撤销；
    - 签发新 Refresh Token 并持久化，响应返回新 refresh_token（Token 轮换）。

    v1.3 增强（N-2）：
    - 旧 Token 撤销由「先 SELECT 再 UPDATE」改为条件 UPDATE
      （`WHERE revoked = 0`），`rowcount == 0` 时回滚并返回 10001，
      避免并发刷新导致同一 Token 对应多个新 Token。

    v1.3.2 修复（MissingGreenlet）：
    - 在 `db.execute(update(...))` 之前，把所有需要的 ORM 数据提取为普通
      Python 值（`user_info` / `original_ip` / `original_user_agent`）。
      SQLAlchemy 2.0 执行 UPDATE 语句后可能使 session 内的 ORM 对象属性
      进入过期状态，之后访问这些属性会触发同步懒加载，在 asyncio 上下文
      中抛出 MissingGreenlet → HTTP 500。
    """
    # 1. 解码
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

    # 2. 只读预检查
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

    # P1-3：纵深防御，校验 Token 载荷与记录一致
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

    # ★ v1.3.2 关键修复：
    # 在 db.execute(update(...)) 之前将所有需要的 ORM 数据提取为普通 Python 值。
    # 这样 UPDATE 之后就不再访问任何 ORM 对象属性，避免 SQLAlchemy 因对象过期
    # 触发同步懒加载，导致 MissingGreenlet → HTTP 500。
    user_info = serialize_user(user)
    original_ip = record.ip
    original_user_agent = record.user_agent

    # 3. 原子撤销旧 Token（N-2：条件 UPDATE 保证并发安全）
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
        # 已被其他并发请求抢先撤销，放弃本次刷新
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

    # 4. 签发新的 Access Token 与 Refresh Token
    #    仅使用已提取的普通 Python 数据，不再触碰 ORM 对象
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

    v1.3（N-6）：改为条件 UPDATE，单条 SQL 完成撤销。
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


async def revoke_all_user_tokens(db: AsyncSession, user_id: int) -> int:
    """
    撤销某用户的全部 Refresh Token，返回撤销数量。

    v1.3（N-6）：改为批量 UPDATE，避免逐条更新带来的 N 次往返。
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

    # 修改密码后撤销该用户全部 Refresh Token
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
    # 平台核心
    {"code": "dashboard:view", "name": "查看仪表盘", "module_id": "platform", "resource": "dashboard", "action": "view"},
    # 认证 - 用户
    {"code": "auth:user:view", "name": "查看用户", "module_id": "auth", "resource": "user", "action": "view"},
    {"code": "auth:user:create", "name": "创建用户", "module_id": "auth", "resource": "user", "action": "create"},
    {"code": "auth:user:edit", "name": "编辑用户", "module_id": "auth", "resource": "user", "action": "edit"},
    {"code": "auth:user:delete", "name": "删除用户", "module_id": "auth", "resource": "user", "action": "delete"},
    # 认证 - 角色
    {"code": "auth:role:view", "name": "查看角色", "module_id": "auth", "resource": "role", "action": "view"},
    {"code": "auth:role:create", "name": "创建角色", "module_id": "auth", "resource": "role", "action": "create"},
    {"code": "auth:role:edit", "name": "编辑角色", "module_id": "auth", "resource": "role", "action": "edit"},
    {"code": "auth:role:delete", "name": "删除角色", "module_id": "auth", "resource": "role", "action": "delete"},
    # 认证 - 权限
    {"code": "auth:permission:view", "name": "查看权限", "module_id": "auth", "resource": "permission", "action": "view"},
    # 模块管理
    {"code": "module_manager:module:view", "name": "查看模块", "module_id": "module_manager", "resource": "module", "action": "view"},
    {"code": "module_manager:module:create", "name": "安装模块", "module_id": "module_manager", "resource": "module", "action": "create"},
    {"code": "module_manager:module:edit", "name": "编辑模块", "module_id": "module_manager", "resource": "module", "action": "edit"},
    {"code": "module_manager:module:delete", "name": "卸载模块", "module_id": "module_manager", "resource": "module", "action": "delete"},
    # 日志审计
    {"code": "audit_log:log:view", "name": "查看日志", "module_id": "audit_log", "resource": "log", "action": "view"},
    {"code": "audit_log:log:export", "name": "导出日志", "module_id": "audit_log", "resource": "log", "action": "export"},
    # License
    {"code": "license:license:view", "name": "查看License", "module_id": "license", "resource": "license", "action": "view"},
    {"code": "license:license:create", "name": "管理License", "module_id": "license", "resource": "license", "action": "create"},
]

# guest 角色的确定性初始权限集合（仅在 guest 角色新建时应用）
GUEST_PERMISSION_CODES = {
    "dashboard:view",
    "auth:user:view",
    "module_manager:module:view",
}


async def _upsert_permissions(
    db: AsyncSession,
) -> Dict[str, Permission]:
    """
    按 code upsert 核心权限。不更新已存在权限的元信息（保持模块注册的权威性）。
    """
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
    """按 code 查询角色，预加载 permissions，避免异步懒加载。"""
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
    """
    Upsert 系统角色：

    - 角色不存在：创建并分配 initial_permissions；
    - 角色已存在：**仅追加**缺失的权限，不删除已有权限。

    ⚠️ v1.3.1 修复要点：
        新建角色时必须在 `db.add()` / `flush()` **之前** 设置
        `role.permissions`。因为：
        - `db.add()` 前：role 处于 transient/pending 状态，属性赋值是纯内存操作；
        - `db.add()` + `flush()` 后：role 变为 persistent 状态，其
          relationship 未加载时会触发 SQLAlchemy 同步懒加载，在 asyncio
          上下文里会抛出 `MissingGreenlet: greenlet_spawn has not been called`。
    """
    role = await _get_role_with_permissions(db, code)
    if not role:
        role = Role(
            name=name,
            code=code,
            description=description,
            is_system=1,
        )
        # ✅ 关键修复：在 db.add() 之前赋值 secondary 关系
        role.permissions = list(initial_permissions)
        db.add(role)
        await db.flush()
        return role

    # 已存在：增量补充
    # 此时 role.permissions 已通过 selectinload 预加载，可以直接遍历
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
    """
    Upsert 默认用户，已存在则复用（不修改密码）。

    ⚠️ v1.3.1 修复要点：
        这里**不再**通过 selectinload 预加载 roles。
        后续用户-角色关联判断改为直接查询 `auth_user_role` 中间表
        （见 `_ensure_user_role`），避免访问未预加载的 relationship。
    """
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
    """
    幂等地确保 user_id 与 role_id 存在关联。

    ⚠️ v1.3.1 修复要点：
        直接查询 `auth_user_role` 中间表，而不是访问 `user.roles`。
        因为 `_upsert_user` 新建用户后，user.roles 关系未预加载，
        访问 `user.roles` 会触发 SQLAlchemy 同步懒加载，
        在 asyncio 上下文中抛出 `MissingGreenlet`。
    """
    exists = await db.scalar(
        select(UserRole).where(
            UserRole.user_id == user_id,
            UserRole.role_id == role_id,
        )
    )
    if not exists:
        db.add(UserRole(user_id=user_id, role_id=role_id))


async def ensure_auth_seed_data(db: AsyncSession) -> None:
    """
    初始化认证模块种子数据（幂等、增量、不破坏已授权权限）。

    该函数可被多次调用：

    - 权限：按 code upsert，不修改已存在权限；
    - 角色：不存在则创建并分配初始权限；存在则仅追加缺失的核心权限；
    - 用户：按 username upsert，已存在则复用；
    - 用户-角色：不存在则追加。

    关键约束：admin 角色在多次启动后仍保留所有已注册权限，
    包括模块安装时动态注册的业务权限。

    ⚠️ v1.3.1 修复要点：
        用户-角色关联改为通过 `_ensure_user_role` 直接操作中间表，
        避免访问未预加载的 `user.roles` 关系，消除 MissingGreenlet。
    """
    # 1. Upsert 核心权限
    permissions_by_code = await _upsert_permissions(db)

    # 2. Upsert 系统角色（增量补充，不删除已有权限）
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

    # 3. Upsert 默认用户
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

    # 4. 追加用户-角色关联（幂等）
    # ✅ 修复：通过 UserRole 中间表查询，避免访问 user.roles 触发 lazy load
    await _ensure_user_role(db, admin_user.id, admin_role.id)
    await _ensure_user_role(db, guest_user.id, guest_role.id)

    await db.commit()
    logger.info("认证种子数据初始化完成（幂等、增量）")
