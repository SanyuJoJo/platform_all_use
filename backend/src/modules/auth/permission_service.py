"""
权限管理模块 - 业务逻辑（v1.3）

v1.3 新增（P2-NEW-6 配套）：
- 新增 audit_permission_format 巡检函数，用于 CI/运维检查权限编码格式合规性。

v1.2 修复（保留）：
- P2-NEW-1：list_permissions 增加 module_id 格式校验。
- P2-NEW-2：register_permissions 失败路径补充 PERMISSION_EVENT ... status=fail 日志。
- P2-NEW-4：register_permissions 数据库操作阶段使用 async with db.begin_nested()
        建立 savepoint，异常时自动回滚，保证 session 可用。
- P2-NEW-5：perm.module_id != module_id 分支明确标注为防御性代码。

v1.1 修复（保留）：
- P0-1：register_permissions 事务原子性修复。
- P0-2：_validate_permission_item 增加 isinstance(value, str) 类型校验。
- P0-3：重复 code 视为参数错误，返回 90001（HTTP 400）。
- P1-1：模块 ID 正则改为 ^[a-z][a-z0-9_]*$。
- P1-2：权限编码正则改为 ^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$。
- P1-6：register_permissions 增加结构化日志记录。
- P1-7：已存在且同模块时仅更新 name。
- P2-4：返回 total 改为 created + updated。
- P2-5：get_permissions_by_module 增加 module_id 格式校验。
"""

import logging
import re
from typing import Any, Dict, List, Optional

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.exceptions import PlatformException
from src.modules.auth.models import Permission

logger = logging.getLogger(__name__)

# 权限编码：{module}:{resource}:{action}，v1.1 起强制小写
_PERMISSION_CODE_PATTERN = re.compile(
    r"^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$"
)
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
    """校验模块 ID（v1.1：强制小写；v1.1：类型校验）。"""
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
    """校验单个权限注册项（v1.1 增加类型校验与小写编码）。"""
    if not isinstance(item, dict):
        raise PlatformException(
            code=90001,
            message="权限项必须为对象",
            status_code=400,
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
            code=20054,
            message=f"权限编码格式无效：{code}",
            status_code=400,
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
    """查询权限列表。v1.2（P2-NEW-1）：module_id 提供时进行格式校验。"""
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
    db: AsyncSession,
    module_id: str,
) -> List[Dict[str, Any]]:
    """获取指定模块的所有权限。v1.1（P2-5）：增加 module_id 格式校验。"""
    _validate_module_id(module_id)
    return await list_permissions(db, module_id=module_id)


async def audit_permission_format(db: AsyncSession) -> Dict[str, Any]:
    """
    ★ v1.3 新增（P2-NEW-6 配套）：巡检所有权限编码格式。

    用于 CI 或运维检查权限编码是否全部符合
    `{module}:{resource}:{action}` 三段格式。

    返回：
        {
            "total": 权限总数,
            "invalid_count": 非法权限数量,
            "invalid": [{"id": ..., "code": ..., "module_id": ...}, ...]
        }
    """
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
    """注册模块权限（供模块管理安装模块时调用）。完整说明见 v1.2 文档。"""
    _validate_module_id(module_id)

    if permissions is None:
        permissions = []

    if not isinstance(permissions, list):
        raise PlatformException(
            code=90001,
            message="permissions 必须为数组",
            status_code=400,
        )

    normalized: List[Dict[str, Any]] = []
    seen = set()

    for item in permissions:
        _validate_permission_item(module_id, item)
        code = item["code"]

        if code in seen:
            raise PlatformException(
                code=90001,
                message=f"权限编码重复：{code}",
                status_code=400,
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
                    # P2-NEW-5：防御性代码。正常 API 路径不可达。
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
