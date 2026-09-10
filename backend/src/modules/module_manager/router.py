"""模块管理模块 - 路由定义（mock 版）。"""

from typing import Optional

from fastapi import APIRouter, Depends, Query

from src.core.response import success_response
from src.modules.auth.dependencies import CurrentUser, require_permission
from .service import get_modules

router = APIRouter(prefix="/api/v1/modules", tags=["Module Manager"])


@router.get("")
async def list_modules_endpoint(
    status: Optional[str] = Query(None, pattern="^(active|inactive)$"),
    keyword: Optional[str] = None,
    current_user: CurrentUser = Depends(
        require_permission("module_manager:module:view")
    ),
):
    """
    查询模块列表（含菜单权限过滤）。

    - 权限：`module_manager:module:view`
    - 返回格式遵循 API 文档 § 1.3 分页格式
    - 每个模块的 menus 字段已按当前用户权限过滤
    """
    user_permissions = current_user.get("permissions", [])
    modules_data = get_modules(status, keyword, user_permissions)
    total = len(modules_data)
    return success_response(
        data={
            "items": modules_data,
            "total": total,
            "page": 1,
            "page_size": total if total > 0 else 1,
            "pages": 1,
        }
    )
