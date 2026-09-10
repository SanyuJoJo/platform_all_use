"""
角色管理模块 - 业务逻辑（v1.3.1）

v1.3.1 Bug 修复：
- 修复 delete_role 删除角色时返回 500 的问题：
  根因：先手动执行 `delete(RolePermission).where(role_id == role_id)` 清空
        secondary 表，随后 `db.delete(role)` 触发 SQLAlchemy 再次尝试删除
        已不存在的 secondary 行（ORM 缓存中仍持有预加载的 permissions 引用），
        抛出 StaleDataError / ObjectDeletedError → HTTP 500。
  修复：移除手动 `delete(RolePermission)`，改为直接 `db.delete(role)`，
        由 SQLAlchemy 通过 `Role.permissions` 的 `secondary="auth_role_permission"`
        关系自动清理中间表。与用户管理模块保持一致（secondary 关系由 ORM 维护）。

v1.3 修复（保留）：
- P2-NEW-C：删除 _map_role_integrity_error 兜底分支中冗余的
        ix_auth_role_code / uq_auth_role_code 条件。

v1.2 修复（保留）：
- P1-NEW-1：delete_role 在 db.delete(role) 之前提取 role_code / role_name。
- P1-NEW-2：重写 _map_role_integrity_error 三级匹配策略（SQLite/PG）。
- P1-NEW-3：新增 _DESCRIPTION_MAX 与 _validate_description。
- P2-NEW-3：delete_role 的 commit() 包裹 try/except IntegrityError + rollback。
- P2-NEW-4：_validate_role_name 签名改为 name: str。

v1.1 修复（保留）：
- P0-1 ~ P0-6、P1-2 ~ P1-7、P2-3、P2-6 全部保留。

防 MissingGreenlet 设计：
- 所有 relationship 使用 selectinload 预加载后再访问；
- 执行 DELETE / UPDATE 后通过 populate_existing=True 重新加载；
- delete 之前提取所有需要的普通 Python 数据。
"""
import re
from typing import Any, Dict, List, Optional

from sqlalchemy import delete, func, or_, select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from src.core.exceptions import PlatformException
from src.modules.auth.models import (
    Permission,
    Role,
    RolePermission,
    UserRole,
)
from src.modules.auth.service import log_auth_event

# ---------------------------------------------------------------------------
# 常量
# ---------------------------------------------------------------------------
_ROLE_NAME_MIN = 1
_ROLE_NAME_MAX = 50
_ROLE_CODE_MIN = 1
_ROLE_CODE_MAX = 50
_DESCRIPTION_MAX = 255

# 角色编码：小写字母开头，后跟小写字母 / 数字 / 下划线
_ROLE_CODE_PATTERN = re.compile(r"^[a-z][a-z0-9_]*$")
# 权限编码：{module}:{resource}:{action}
_PERMISSION_CODE_PATTERN = re.compile(
    r"^[a-zA-Z0-9_]+:[a-zA-Z0-9_]+:[a-zA-Z0-9_]+$"
)


# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
async def _load_role_with_permissions(
    db: AsyncSession,
    role_id: int,
    *,
    populate_existing: bool = False,
) -> Optional[Role]:
    """按 ID 加载角色并预加载权限。"""
    stmt = (
        select(Role)
        .options(selectinload(Role.permissions))
        .where(Role.id == role_id)
    )
    if populate_existing:
        stmt = stmt.execution_options(populate_existing=True)
    return await db.scalar(stmt)


def _serialize_role(role: Role) -> Dict[str, Any]:
    """将 Role ORM 对象序列化为统一返回字典。"""
    return {
        "id": role.id,
        "name": role.name,
        "code": role.code,
        "description": role.description,
        "is_system": role.is_system,
        "permission_codes": sorted([p.code for p in role.permissions]),
        "created_at": role.created_at,
        "updated_at": role.updated_at,
    }


def _validate_role_name(name: str) -> None:
    """校验角色名称长度，非法时抛出 90001。"""
    if not (_ROLE_NAME_MIN <= len(name) <= _ROLE_NAME_MAX):
        raise PlatformException(
            code=90001,
            message=f"角色名称长度必须为 {_ROLE_NAME_MIN}-{_ROLE_NAME_MAX} 位",
            status_code=400,
        )


def _validate_role_code(code: str) -> None:
    """校验角色编码长度与格式，非法时抛出 90001。"""
    if not (_ROLE_CODE_MIN <= len(code) <= _ROLE_CODE_MAX):
        raise PlatformException(
            code=90001,
            message=f"角色编码长度必须为 {_ROLE_CODE_MIN}-{_ROLE_CODE_MAX} 位",
            status_code=400,
        )
    if not _ROLE_CODE_PATTERN.match(code):
        raise PlatformException(
            code=90001,
            message="角色编码格式无效（小写字母开头，仅允许小写字母、数字、下划线）",
            status_code=400,
        )


def _validate_description(description: Optional[str]) -> None:
    """校验角色描述长度，非法时抛出 90001。"""
    if description is None:
        return
    if len(description) > _DESCRIPTION_MAX:
        raise PlatformException(
            code=90001,
            message=f"角色描述长度不得超过 {_DESCRIPTION_MAX} 位",
            status_code=400,
        )


def _normalize_permission_codes(codes: Optional[List[str]]) -> List[str]:
    """权限编码保序去重。"""
    if not codes:
        return []
    seen = set()
    result: List[str] = []
    for code in codes:
        if code not in seen:
            seen.add(code)
            result.append(code)
    return result


def _validate_permission_codes_format(codes: List[str]) -> None:
    """校验权限编码格式，非法时抛出 20054。"""
    for code in codes:
        if not _PERMISSION_CODE_PATTERN.match(code):
            raise PlatformException(
                code=20054,
                message=f"权限编码格式无效：{code}",
                status_code=400,
            )


async def _load_permissions_by_codes(
    db: AsyncSession,
    codes: List[str],
) -> List[Permission]:
    """按权限编码加载 Permission 对象，缺失时抛出 20052。"""
    if not codes:
        return []

    result = await db.execute(
        select(Permission).where(Permission.code.in_(codes))
    )
    permissions = list(result.scalars().all())
    permission_map = {p.code: p for p in permissions}

    missing = [code for code in codes if code not in permission_map]
    if missing:
        raise PlatformException(
            code=20052,
            message=f"权限不存在：{missing}",
            status_code=404,
        )

    return [permission_map[code] for code in codes]


def _map_role_integrity_error(exc: IntegrityError) -> PlatformException:
    """将角色唯一约束冲突映射为业务错误码。

    三级匹配策略（v1.3 精简）：
        1. 从 exc.orig 中提取约束名（正则）；
        2. 按约束名中的关键字判断：role_code / auth_role.code；
        3. 兜底：auth_role.code / auth_role_code。
    """
    orig = getattr(exc, "orig", exc)
    orig_str = str(orig)
    lower = orig_str.lower()

    # ---- 1. 提取约束名 ----
    constraint_names: List[str] = []
    constraint_names.extend(
        re.findall(r'unique constraint\s+"([^"]+)"', lower)
    )
    constraint_names.extend(
        re.findall(r"unique constraint failed:\s*([a-z0-9_.]+)", lower)
    )
    constraint_names.extend(
        re.findall(r"\b(ix_[a-z0-9_]+|uq_[a-z0-9_]+)\b", lower)
    )
    combined = " ".join(constraint_names)

    # ---- 2. 按约束名匹配 ----
    if (
        "role_code" in combined
        or "auth_role.code" in combined
        or "auth_role_code" in combined
    ):
        return PlatformException(
            code=20001,
            message="角色编码已存在",
            status_code=400,
        )

    # ---- 3. 兜底匹配 ----
    if "auth_role.code" in lower or "auth_role_code" in lower:
        return PlatformException(
            code=20001,
            message="角色编码已存在",
            status_code=400,
        )

    return PlatformException(
        code=90003,
        message="数据冲突",
        status_code=409,
    )


# ---------------------------------------------------------------------------
# 查询列表
# ---------------------------------------------------------------------------
async def list_roles(
    db: AsyncSession,
    page: int,
    page_size: int,
    keyword: Optional[str] = None,
) -> Dict[str, Any]:
    """分页查询角色列表。"""
    conditions = []
    if keyword:
        like = f"%{keyword}%"
        conditions.append(
            or_(
                Role.name.ilike(like),
                Role.code.ilike(like),
            )
        )

    count_stmt = select(func.count(Role.id))
    if conditions:
        count_stmt = count_stmt.where(*conditions)
    total = (await db.execute(count_stmt)).scalar_one()

    list_stmt = select(Role).options(selectinload(Role.permissions))
    if conditions:
        list_stmt = list_stmt.where(*conditions)

    list_stmt = (
        list_stmt.order_by(Role.id.asc())
        .offset((page - 1) * page_size)
        .limit(page_size)
    )
    result = await db.execute(list_stmt)
    roles = list(result.scalars().all())

    return {
        "items": [_serialize_role(role) for role in roles],
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }


# ---------------------------------------------------------------------------
# 详情
# ---------------------------------------------------------------------------
async def get_role_detail(db: AsyncSession, role_id: int) -> Dict[str, Any]:
    """获取角色详情，不存在返回 20002。"""
    role = await _load_role_with_permissions(db, role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)
    return _serialize_role(role)


# ---------------------------------------------------------------------------
# 创建
# ---------------------------------------------------------------------------
async def create_role(
    db: AsyncSession,
    data: Dict[str, Any],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """创建角色。"""
    name: str = data["name"]
    code: str = data["code"]
    description: Optional[str] = data.get("description")
    permission_codes = _normalize_permission_codes(
        data.get("permission_codes") or []
    )

    _validate_role_name(name)
    _validate_role_code(code)
    _validate_description(description)
    _validate_permission_codes_format(permission_codes)

    existing_name = await db.scalar(select(Role.id).where(Role.name == name))
    if existing_name:
        log_auth_event(
            "role_create",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=20004,
            detail=f"角色名称已存在：{name}",
        )
        raise PlatformException(
            code=20004, message="角色名称已存在", status_code=400
        )

    existing_code = await db.scalar(select(Role.id).where(Role.code == code))
    if existing_code:
        log_auth_event(
            "role_create",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=20001,
            detail=f"角色编码已存在：{code}",
        )
        raise PlatformException(
            code=20001, message="角色编码已存在", status_code=400
        )

    permissions = await _load_permissions_by_codes(db, permission_codes)

    role = Role(
        name=name,
        code=code,
        description=description,
        is_system=0,
    )
    # 直接通过 secondary 关系赋值，交由 ORM 插入 auth_role_permission
    role.permissions = permissions
    db.add(role)

    try:
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_role_integrity_error(exc) from exc

    role = await _load_role_with_permissions(db, role.id, populate_existing=True)

    log_auth_event(
        "role_create",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"创建角色 {name}（id={role.id}, code={code}）",
    )
    return _serialize_role(role)


# ---------------------------------------------------------------------------
# 更新
# ---------------------------------------------------------------------------
async def update_role(
    db: AsyncSession,
    role_id: int,
    data: Dict[str, Any],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """更新角色信息，permission_codes 全量覆盖。"""
    role = await _load_role_with_permissions(db, role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)

    if "name" in data and data["name"] is not None:
        _validate_role_name(data["name"])
        existing = await db.scalar(
            select(Role.id).where(
                Role.name == data["name"],
                Role.id != role_id,
            )
        )
        if existing:
            raise PlatformException(
                code=20004, message="角色名称已存在", status_code=400
            )
        role.name = data["name"]

    if "description" in data:
        _validate_description(data["description"])
        role.description = data["description"]

    if "permission_codes" in data and data["permission_codes"] is not None:
        codes = _normalize_permission_codes(data["permission_codes"])
        _validate_permission_codes_format(codes)
        permissions = await _load_permissions_by_codes(db, codes)

        # 通过 secondary 关系全量覆盖，交由 ORM 管理中间表
        role.permissions = permissions

    try:
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_role_integrity_error(exc) from exc

    role = await _load_role_with_permissions(db, role_id, populate_existing=True)

    log_auth_event(
        "role_update",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"更新角色 id={role_id}（code={role.code}）",
    )
    return _serialize_role(role)


# ---------------------------------------------------------------------------
# 删除（v1.3.1 修复）
# ---------------------------------------------------------------------------
async def delete_role(
    db: AsyncSession,
    role_id: int,
    operator: Dict[str, Any],
) -> None:
    """删除角色。

    v1.3.1 修复：
        不再手动 `delete(RolePermission)`。Role.permissions 是
        `secondary="auth_role_permission"` 的关系，SQLAlchemy 在
        `db.delete(role)` 时会自动清理中间表。若先手动删除 secondary
        记录再 delete(role)，ORM 缓存中的 permission 引用会导致
        SQLAlchemy 再次尝试删除已不存在的行，触发 StaleDataError →
        HTTP 500。

    执行顺序：
        1. 角色不存在 → 20002
        2. 系统内置角色 → 20003
        3. 查询 auth_user_role 关联数量 > 0 → 20005
        4. 提取 role_code / role_name 为局部变量
        5. `db.delete(role)`（ORM 自动清理 auth_role_permission）
        6. commit()，包裹 try/except IntegrityError + rollback
        7. 审计日志（使用局部变量）
    """
    role = await _load_role_with_permissions(db, role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)

    if role.is_system == 1:
        raise PlatformException(
            code=20003,
            message="不能删除系统内置角色",
            status_code=403,
        )

    user_count = await db.scalar(
        select(func.count(UserRole.id)).where(UserRole.role_id == role_id)
    )
    if user_count and user_count > 0:
        log_auth_event(
            "role_delete",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=20005,
            detail=f"角色被 {user_count} 个用户使用，id={role_id}",
        )
        raise PlatformException(
            code=20005,
            message="角色已被用户使用，请先解除关联",
            status_code=409,
        )

    # 提取普通 Python 值（P1-NEW-1）
    role_code = role.code
    role_name = role.name

    # ★ v1.3.1 关键修复：
    # 直接 db.delete(role)，由 SQLAlchemy 通过 secondary 关系自动清理
    # auth_role_permission 中间表。不再手动 delete(RolePermission)，
    # 避免 ORM 缓存中预加载的 permissions 引用触发二次删除。
    await db.delete(role)

    try:
        await db.commit()
    except IntegrityError as exc:
        await db.rollback()
        raise _map_role_integrity_error(exc) from exc

    log_auth_event(
        "role_delete",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"删除角色 id={role_id}（code={role_code}, name={role_name}）",
    )
