"""
用户管理模块 - 业务逻辑（v1.3）
v1.3 修复：
- P2-1：reset_user_password 中密码修改与 Refresh Token 撤销在同一事务提交
- P2-4：重写 _map_integrity_error，兼容 SQLite / PostgreSQL / 命名约束
v1.2 修复（保留）：
- P1-2：delete_user 中调用 revoke_all_user_tokens(commit=False)
- P2-2：_map_integrity_error 三级匹配（v1.3 进一步强化）
- P2-4：_validate_nickname 由 Service 层统一返回 90001
v1.1 修复（保留）：
- P1-5：_normalize_email 邮箱归一化
- P1-6：捕获 IntegrityError
- P2-7：_normalize_role_ids 保序去重
- P2-8：keyword 匹配 email 时显式排除 NULL
防 MissingGreenlet 设计（延续认证模块 v1.3.2）：
- 所有 relationship 使用 selectinload 预加载后再访问；
- 执行 DELETE / UPDATE 语句后通过 populate_existing=True 强制刷新
  identity map，避免读取到过期的 relationship 数据。
"""
import re
from typing import Any, Dict, List, Optional
from sqlalchemy import delete, func, or_, select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from src.core.exceptions import PlatformException
from src.core.security import hash_password
from src.modules.auth.models import Role, User, UserRole
from src.modules.auth.service import log_auth_event, revoke_all_user_tokens
# 用户名格式：3-20 位，字母 / 数字 / 下划线
_USERNAME_PATTERN = re.compile(r"^[a-zA-Z0-9_]{3,20}$")
# 密码长度约束
_PASSWORD_MIN = 6
_PASSWORD_MAX = 20
# 昵称长度约束
_NICKNAME_MIN = 1
_NICKNAME_MAX = 50
# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
async def _load_user_with_roles(
    db: AsyncSession, user_id: int, *, populate_existing: bool = False
) -> Optional[User]:
    """按 ID 加载用户并预加载角色。"""
    stmt = (
        select(User)
        .options(selectinload(User.roles))
        .where(User.id == user_id)
    )
    if populate_existing:
        stmt = stmt.execution_options(populate_existing=True)
    return await db.scalar(stmt)
def _serialize_user(user: User) -> Dict[str, Any]:
    """将 User ORM 对象序列化为用户管理模块统一返回字典。"""
    return {
        "id": user.id,
        "username": user.username,
        "nickname": user.nickname,
        "email": user.email,
        "avatar": user.avatar,
        "status": user.status,
        "roles": [
            {"id": r.id, "name": r.name, "code": r.code}
            for r in user.roles
        ],
        "created_at": user.created_at,
        "updated_at": user.updated_at,
    }
async def _validate_role_ids(
    db: AsyncSession, role_ids: List[int]
) -> List[Role]:
    """校验角色 ID 是否全部存在，返回 Role 列表。"""
    if not role_ids:
        return []
    result = await db.execute(select(Role).where(Role.id.in_(role_ids)))
    roles = list(result.scalars().all())
    found = {r.id for r in roles}
    missing = [rid for rid in role_ids if rid not in found]
    if missing:
        raise PlatformException(
            code=20002,
            message=f"角色不存在：{missing}",
            status_code=404,
        )
    return roles
def _normalize_role_ids(role_ids: Optional[List[int]]) -> List[int]:
    """对 role_ids 去重，保持首次出现顺序。"""
    if not role_ids:
        return []
    seen = set()
    result = []
    for rid in role_ids:
        if rid not in seen:
            seen.add(rid)
            result.append(rid)
    return result
def _normalize_email(email: Optional[str]) -> Optional[str]:
    """
    邮箱归一化。
    - None → None
    - "" / "   " → None
    - "  foo@bar.com  " → "foo@bar.com"
    """
    if email is None:
        return None
    return email.strip() or None
def _validate_username(username: str) -> None:
    """校验用户名格式，非法时抛出 10006。"""
    if not _USERNAME_PATTERN.match(username):
        raise PlatformException(
            code=10006,
            message="用户名格式无效（3-20 位字母、数字或下划线）",
            status_code=400,
        )
def _validate_password(password: str) -> None:
    """校验密码长度，非法时抛出 10007。"""
    if not (_PASSWORD_MIN <= len(password) <= _PASSWORD_MAX):
        raise PlatformException(
            code=10007,
            message=f"密码长度必须为 {_PASSWORD_MIN}-{_PASSWORD_MAX} 位",
            status_code=400,
        )
def _validate_nickname(nickname: Optional[str]) -> None:
    """校验昵称长度，非法时抛出 90001。"""
    if nickname is None:
        return
    if not (_NICKNAME_MIN <= len(nickname) <= _NICKNAME_MAX):
        raise PlatformException(
            code=90001,
            message=f"昵称长度必须为 {_NICKNAME_MIN}-{_NICKNAME_MAX} 位",
            status_code=400,
        )
def _map_integrity_error(exc: IntegrityError) -> PlatformException:
    """
    v1.3（P2-4）：将数据库唯一约束冲突映射为业务错误码。
    三级匹配策略：
        1. 从 exc.orig 中提取约束名（正则）：
           - PostgreSQL：duplicate key value violates unique constraint "auth_user_username_key"
           - SQLite   ：UNIQUE constraint failed: auth_user.username
           - 命名约束 ：uq_auth_user_role_user_role / idx_auth_user_email
        2. 按约束名中的关键字判断：user_role / username / email
        3. 兜底：直接在原始报错字符串中按 table.column 匹配
    映射关系：
        - 命中 user_role 相关约束 → 90003
        - 命中 username 相关约束 → 10000
        - 命中 email 相关约束    → 10009
        - 其他冲突               → 90003
    """
    orig = getattr(exc, "orig", exc)
    orig_str = str(orig)
    lower = orig_str.lower()
    # ---- 1. 提取约束名 ----
    constraint_names: List[str] = []
    # PostgreSQL：duplicate key value violates unique constraint "auth_user_username_key"
    constraint_names.extend(
        re.findall(r'unique constraint\s+"([^"]+)"', lower)
    )
    # SQLite：UNIQUE constraint failed: auth_user.username
    constraint_names.extend(
        re.findall(r"unique constraint failed:\s*([a-z0-9_.]+)", lower)
    )
    # 命名约束：uq_* / idx_*
    constraint_names.extend(
        re.findall(r"\b(uq_[a-z0-9_]+|idx_[a-z0-9_]+)\b", lower)
    )
    combined = " ".join(constraint_names)
    # ---- 2. 按约束名匹配 ----
    if "user_role" in combined:
        return PlatformException(
            code=90003, message="用户角色关联冲突", status_code=409
        )
    if "username" in combined:
        return PlatformException(
            code=10000, message="用户名已存在", status_code=400
        )
    if "email" in combined:
        return PlatformException(
            code=10009, message="邮箱已被使用", status_code=409
        )
    # ---- 3. 兜底匹配 ----
    if "user_role" in lower or "uq_auth_user_role" in lower:
        return PlatformException(
            code=90003, message="用户角色关联冲突", status_code=409
        )
    if "auth_user.username" in lower or "auth_user_username" in lower:
        return PlatformException(
            code=10000, message="用户名已存在", status_code=400
        )
    if "auth_user.email" in lower or "auth_user_email" in lower:
        return PlatformException(
            code=10009, message="邮箱已被使用", status_code=409
        )
    return PlatformException(code=90003, message="数据冲突", status_code=409)
# ---------------------------------------------------------------------------
# 查询列表
# ---------------------------------------------------------------------------
async def list_users(
    db: AsyncSession,
    page: int,
    page_size: int,
    keyword: Optional[str] = None,
    status: Optional[int] = None,
    role_id: Optional[int] = None,
) -> Dict[str, Any]:
    """
    分页查询用户列表。
    - keyword：模糊匹配 username / nickname / email
    - status：精确匹配
    - role_id：按角色筛选（子查询避免 JOIN 重复）
    """
    conditions = []
    if keyword:
        like = f"%{keyword}%"
        conditions.append(
            or_(
                User.username.ilike(like),
                User.nickname.ilike(like),
                (User.email.isnot(None)) & (User.email.ilike(like)),
            )
        )
    if status is not None:
        conditions.append(User.status == status)
    if role_id is not None:
        conditions.append(
            User.id.in_(
                select(UserRole.user_id).where(UserRole.role_id == role_id)
            )
        )
    count_stmt = select(func.count(User.id))
    if conditions:
        count_stmt = count_stmt.where(*conditions)
    total = (await db.execute(count_stmt)).scalar_one()
    list_stmt = select(User).options(selectinload(User.roles))
    if conditions:
        list_stmt = list_stmt.where(*conditions)
    list_stmt = (
        list_stmt.order_by(User.id.asc())
        .offset((page - 1) * page_size)
        .limit(page_size)
    )
    result = await db.execute(list_stmt)
    users = list(result.scalars().all())
    return {
        "items": [_serialize_user(u) for u in users],
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }
# ---------------------------------------------------------------------------
# 详情
# ---------------------------------------------------------------------------
async def get_user_detail(db: AsyncSession, user_id: int) -> Dict[str, Any]:
    """获取用户详情，用户不存在返回 10005。"""
    user = await _load_user_with_roles(db, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    return _serialize_user(user)
# ---------------------------------------------------------------------------
# 创建
# ---------------------------------------------------------------------------
async def create_user(
    db: AsyncSession,
    data: Dict[str, Any],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """
    创建用户。
    校验顺序：
        1. 用户名格式（10006）
        2. 密码长度（10007）
        3. 昵称长度（90001）
        4. 用户名唯一（10000）
        5. 邮箱唯一（10009）
        6. 角色存在性（20002）
    """
    username: str = data["username"]
    password: str = data["password"]
    nickname: str = data["nickname"]
    email = _normalize_email(data.get("email"))
    role_ids = _normalize_role_ids(data.get("role_ids") or [])
    status: int = data.get("status", 1)
    _validate_username(username)
    _validate_password(password)
    _validate_nickname(nickname)
    existing_username = await db.scalar(
        select(User.id).where(User.username == username)
    )
    if existing_username:
        log_auth_event(
            "user_create",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=10000,
            detail=f"用户名已存在：{username}",
        )
        raise PlatformException(code=10000, message="用户名已存在", status_code=400)
    if email:
        existing_email = await db.scalar(
            select(User.id).where(User.email == email)
        )
        if existing_email:
            log_auth_event(
                "user_create",
                user_id=operator["id"],
                username=operator["username"],
                status="fail",
                error_code=10009,
                detail=f"邮箱已被使用：{email}",
            )
            raise PlatformException(code=10009, message="邮箱已被使用", status_code=409)
    roles = await _validate_role_ids(db, role_ids)
    user = User(
        username=username,
        password_hash=hash_password(password),
        nickname=nickname,
        email=email,
        status=status,
    )
    db.add(user)
    try:
        await db.flush()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_integrity_error(exc) from exc
    for role in roles:
        db.add(UserRole(user_id=user.id, role_id=role.id))
    try:
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_integrity_error(exc) from exc
    user = await _load_user_with_roles(db, user.id, populate_existing=True)
    log_auth_event(
        "user_create",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"创建用户 {username}（id={user.id}）",
    )
    return _serialize_user(user)
# ---------------------------------------------------------------------------
# 更新
# ---------------------------------------------------------------------------
async def update_user(
    db: AsyncSession,
    user_id: int,
    data: Dict[str, Any],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """
    更新用户信息。
    - role_ids 全量覆盖
    - 禁用自己返回 10011
    - 邮箱重复返回 10009
    - 昵称超长返回 90001
    """
    user = await _load_user_with_roles(db, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if "nickname" in data and data["nickname"] is not None:
        _validate_nickname(data["nickname"])
        user.nickname = data["nickname"]
    if "email" in data:
        new_email = _normalize_email(data["email"])
        if new_email:
            existing = await db.scalar(
                select(User.id).where(User.email == new_email, User.id != user_id)
            )
            if existing:
                raise PlatformException(
                    code=10009, message="邮箱已被使用", status_code=409
                )
        user.email = new_email
    if "status" in data and data["status"] is not None:
        if user_id == operator["id"] and data["status"] == 0:
            raise PlatformException(
                code=10011, message="不能禁用自己", status_code=403
            )
        user.status = data["status"]
    if "role_ids" in data and data["role_ids"] is not None:
        role_ids = _normalize_role_ids(data["role_ids"])
        roles = await _validate_role_ids(db, role_ids)
        await db.execute(delete(UserRole).where(UserRole.user_id == user_id))
        for role in roles:
            db.add(UserRole(user_id=user_id, role_id=role.id))
    try:
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_integrity_error(exc) from exc
    user = await _load_user_with_roles(db, user_id, populate_existing=True)
    log_auth_event(
        "user_update",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"更新用户 id={user_id}",
    )
    return _serialize_user(user)
# ---------------------------------------------------------------------------
# 删除（v1.2 P1-2 修复保留）
# ---------------------------------------------------------------------------
async def delete_user(
    db: AsyncSession,
    user_id: int,
    operator: Dict[str, Any],
) -> None:
    """
    删除用户。
    - 不能删除自己（10010）
    - 用户不存在（10005）
    - 撤销 Refresh Token 与删除用户在同一事务提交。
    执行顺序（同一事务）：
        1. 撤销该用户所有 Refresh Token（commit=False）
        2. 删除 auth_user_role 关联
        3. 删除 auth_user
        4. db.commit()
    """
    if user_id == operator["id"]:
        raise PlatformException(code=10010, message="不能删除自己", status_code=403)
    user = await db.get(User, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    # 撤销 Refresh Token，但不提交
    await revoke_all_user_tokens(db, user_id, commit=False)
    # 清理用户-角色关联，再删除用户
    await db.execute(delete(UserRole).where(UserRole.user_id == user_id))
    await db.delete(user)
    # 一次性提交全部变更；若前述步骤失败，本次回滚，撤销也会被回滚
    await db.commit()
    log_auth_event(
        "user_delete",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"删除用户 id={user_id}",
    )
# ---------------------------------------------------------------------------
# 启用/禁用
# ---------------------------------------------------------------------------
async def update_user_status(
    db: AsyncSession,
    user_id: int,
    status: int,
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """启用/禁用用户。不能禁用自己（10011）。"""
    if user_id == operator["id"] and status == 0:
        raise PlatformException(code=10011, message="不能禁用自己", status_code=403)
    user = await _load_user_with_roles(db, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    user.status = status
    await db.commit()
    user = await _load_user_with_roles(db, user_id, populate_existing=True)
    log_auth_event(
        "user_status",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"用户 id={user_id} 状态变更为 {status}",
    )
    return _serialize_user(user)
# ---------------------------------------------------------------------------
# 重置密码（v1.3 P2-1 修复）
# ---------------------------------------------------------------------------
async def reset_user_password(
    db: AsyncSession,
    user_id: int,
    new_password: str,
    operator: Dict[str, Any],
) -> None:
    """
    管理员重置用户密码。
    - 密码长度 6-20（10007）
    - 用户不存在（10005）
    - v1.3（P2-1）：密码修改与 Refresh Token 撤销在同一事务提交，
      若撤销失败则密码修改一并回滚。
    执行顺序（同一事务）：
        1. 更新 password_hash（不提交）
        2. 撤销 Refresh Token（commit=False）
        3. db.commit()
    """
    _validate_password(new_password)
    user = await db.get(User, user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    user.password_hash = hash_password(new_password)
    # v1.3（P2-1）：撤销 Refresh Token，但不提交
    await revoke_all_user_tokens(db, user_id, commit=False)
    # 一次性提交密码修改 + Token 撤销
    await db.commit()
    log_auth_event(
        "user_password_reset",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"管理员重置用户 id={user_id}（{user.username}）的密码",
    )
