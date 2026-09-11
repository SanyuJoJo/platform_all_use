"""
权限管理模块 - 业务逻辑（v1.2）

v1.2 变更：
- V11-P0-09：将 delete / RolePermission 的导入从函数内移到文件顶部。
"""
import logging
import re
from typing import Any, Dict, List, Optional

from sqlalchemy import delete, select
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.exceptions import PlatformException
from src.modules.auth.models import Permission, RolePermission

logger = logging.getLogger(__name__)

# 权限编码：{module}:{resource}:{action}
_PERMISSION_CODE_PATTERN = re.compile(r"^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$")
# 模块 ID：小写字母开头，后跟小写字母 / 数字 / 下划线
_MODULE_ID_PATTERN = re.compile(r"^[a-z][a-z0-9_]*$")

_NAME_MAX = 50
_RESOURCE_MAX = 50
_ACTION_MAX = 50


def _serialize_permission(perm: Permission) -> Dict[str, Any]:
    """将 Permission ORM 对象序列化为统一字典。"""
    return {
        "id": perm.id,
        "code": perm.code,
        "name": perm.name,
        "module_id": perm.module_id,
        "resource": perm.resource,
        "action": perm.action,
        "created_at": perm.created_at,
    }


def _validate_module_id(module_id: str) -> None:
    """校验模块 ID。"""
    if not module_id or not isinstance(module_id, str):
        raise PlatformException(
            code=90001,
            message="module_id 必须为非空字符串",
            status_code=400,
        )
    if not (1 <= len(module_id) <= 50) or not _MODULE_ID_PATTERN.match(module_id):
        raise PlatformException(
            code=90001,
            message="module_id 格式无效（小写字母开头，仅允许小写字母、数字、下划线）",
            status_code=400,
        )


def _validate_permission_item(module_id: str, item: Dict[str, Any]) -> None:
    """校验单个权限注册项。"""
    if not isinstance(item, dict):
        raise PlatformException(
            code=90001, message="权限项必须为对象", status_code=400
        )

    for field in ("code", "name", "resource", "action"):
        value = item.get(field)
        if not value or not isinstance(value, str):
            raise PlatformException(
                code=90001,
                message=f"权限字段 {field} 必须为非空字符串",
                status_code=400,
            )

    code = item["code"]

    if not _PERMISSION_CODE_PATTERN.match(code):
        raise PlatformException(
            code=20054,
            message=f"权限编码格式无效（必须为小写 {{module}}:{{resource}}:{{action}}）：{code}",
            status_code=400,
        )

    parts = code.split(":")
    if len(parts) != 3:
        raise PlatformException(
            code=20054, message=f"权限编码格式无效：{code}", status_code=400
        )

    code_module, code_resource, code_action = parts

    if code_module != module_id:
        raise PlatformException(
            code=20054,
            message=f"权限编码模块前缀与 module_id 不一致：{code}",
            status_code=400,
        )

    if code_resource != item["resource"] or code_action != item["action"]:
        raise PlatformException(
            code=20054,
            message=f"权限编码与 resource/action 不一致：{code}",
            status_code=400,
        )

    if not (1 <= len(item["name"]) <= _NAME_MAX):
        raise PlatformException(
            code=90001,
            message=f"权限名称长度必须为 1-{_NAME_MAX} 位",
            status_code=400,
        )

    if not (1 <= len(item["resource"]) <= _RESOURCE_MAX):
        raise PlatformException(
            code=90001,
            message=f"资源名称长度必须为 1-{_RESOURCE_MAX} 位",
            status_code=400,
        )

    if not (1 <= len(item["action"]) <= _ACTION_MAX):
        raise PlatformException(
            code=90001,
            message=f"操作名称长度必须为 1-{_ACTION_MAX} 位",
            status_code=400,
        )


async def list_permissions(
    db: AsyncSession,
    module_id: Optional[str] = None,
    resource: Optional[str] = None,
) -> List[Dict[str, Any]]:
    """查询权限列表。"""
    if module_id is not None and module_id != "":
        _validate_module_id(module_id)

    stmt = select(Permission)

    if module_id:
        stmt = stmt.where(Permission.module_id == module_id)
    if resource:
        stmt = stmt.where(Permission.resource == resource)

    stmt = stmt.order_by(
        Permission.module_id.asc(),
        Permission.resource.asc(),
        Permission.action.asc(),
        Permission.id.asc(),
    )

    result = await db.execute(stmt)
    return [_serialize_permission(p) for p in result.scalars().all()]


async def get_permissions_by_module(
    db: AsyncSession, module_id: str
) -> List[Dict[str, Any]]:
    """获取指定模块的所有权限。"""
    _validate_module_id(module_id)
    return await list_permissions(db, module_id=module_id)


async def audit_permission_format(db: AsyncSession) -> Dict[str, Any]:
    """巡检所有权限编码格式。"""
    result = await db.execute(select(Permission))
    all_perms = list(result.scalars().all())

    invalid = [
        {
            "id": p.id,
            "code": p.code,
            "module_id": p.module_id,
            "resource": p.resource,
            "action": p.action,
        }
        for p in all_perms
        if not _PERMISSION_CODE_PATTERN.match(p.code)
    ]

    return {
        "total": len(all_perms),
        "invalid_count": len(invalid),
        "invalid": invalid,
    }


async def register_permissions(
    db: AsyncSession,
    module_id: str,
    permissions: List[Dict[str, Any]],
    *,
    commit: bool = True,
) -> Dict[str, int]:
    """注册模块权限（供模块管理安装模块时调用）。"""
    _validate_module_id(module_id)

    if permissions is None:
        permissions = []

    if not isinstance(permissions, list):
        raise PlatformException(
            code=90001, message="permissions 必须为数组", status_code=400
        )

    normalized: List[Dict[str, Any]] = []
    seen = set()

    for item in permissions:
        _validate_permission_item(module_id, item)
        code = item["code"]

        if code in seen:
            raise PlatformException(
                code=90001, message=f"权限编码重复：{code}", status_code=400
            )
        seen.add(code)

        normalized.append(
            {
                "code": code,
                "name": item["name"],
                "resource": item["resource"],
                "action": item["action"],
            }
        )

    created = 0
    updated = 0

    try:
        async with db.begin_nested():
            for item in normalized:
                perm = await db.scalar(
                    select(Permission).where(Permission.code == item["code"])
                )

                if perm:
                    if perm.module_id != module_id:
                        raise PlatformException(
                            code=20053,
                            message=(
                                f"权限编码已被模块 {perm.module_id} 占用："
                                f"{item['code']}"
                            ),
                            status_code=409,
                        )
                    perm.name = item["name"]
                    updated += 1
                else:
                    perm = Permission(
                        code=item["code"],
                        name=item["name"],
                        module_id=module_id,
                        resource=item["resource"],
                        action=item["action"],
                    )
                    db.add(perm)
                    created += 1

        if commit:
            await db.commit()

    except Exception as exc:
        if commit:
            await db.rollback()

        error_code = getattr(exc, "code", "unknown")
        logger.error(
            "PERMISSION_EVENT action=register_permissions module_id=%s "
            "status=fail error_code=%s detail=%s",
            module_id,
            error_code,
            str(exc)[:200],
        )
        raise

    logger.info(
        "PERMISSION_EVENT action=register_permissions module_id=%s "
        "created=%s updated=%s total=%s status=success",
        module_id,
        created,
        updated,
        created + updated,
    )

    return {
        "created": created,
        "updated": updated,
        "total": created + updated,
    }


# ---------------------------------------------------------------------------
# v1.1 新增：权限卸载（v1.2：导入顶部化）
# ---------------------------------------------------------------------------
async def unregister_permissions(
    db: AsyncSession,
    module_id: str,
    *,
    commit: bool = True,
) -> int:
    """
    卸载模块时清理该模块注册的所有权限。

    执行步骤：
        1. 查询该模块下所有 Permission.id；
        2. 删除这些权限对应的 auth_role_permission 关联；
        3. 删除 auth_permission 记录；
        4. commit（可选）。
    """
    _validate_module_id(module_id)

    result = await db.execute(
        select(Permission.id).where(Permission.module_id == module_id)
    )
    perm_ids = [row[0] for row in result.all()]

    if not perm_ids:
        logger.info(
            "PERMISSION_EVENT action=unregister_permissions "
            "module_id=%s deleted=0 status=success",
            module_id,
        )
        return 0

    await db.execute(
        delete(RolePermission).where(RolePermission.permission_id.in_(perm_ids))
    )
    delete_result = await db.execute(
        delete(Permission).where(Permission.id.in_(perm_ids))
    )

    if commit:
        await db.commit()

    deleted = delete_result.rowcount or 0
    logger.info(
        "PERMISSION_EVENT action=unregister_permissions "
        "module_id=%s deleted=%s status=success",
        module_id,
        deleted,
    )
    return deleted
