"""
权限管理模块 - 路由定义（v1.1，v1.2 未修改）
路由前缀：/api/v1/auth/permissions
权限点：auth:permission:view
v1.1 变更：
- P1-5：路由函数重命名，避免与 service 函数同名。
"""
from typing import Optional
from fastapi import APIRouter, Depends, Path, Query
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.database import get_db
from src.core.response import success_response
from src.modules.auth import permission_service
from src.modules.auth.dependencies import CurrentUser, require_permission
router = APIRouter(prefix="/api/v1/auth/permissions", tags=["Permission"])
@router.get("")
async def list_permissions_endpoint(
    module_id: Optional[str] = Query(None, description="按模块筛选"),
    resource: Optional[str] = Query(None, description="按资源筛选"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:permission:view")),
):
    """
    查询权限列表。
    - 权限：`auth:permission:view`
    - 支持按 `module_id`、`resource` 筛选
    - 返回数组，不分页，符合 API 文档 § 五
    - v1.2：module_id 提供时进行格式校验，非法返回 90001
    """
    data = await permission_service.list_permissions(
        db,
        module_id=module_id,
        resource=resource,
    )
    return success_response(data=data)
@router.get("/modules/{module_id}")
async def get_module_permissions_endpoint(
    module_id: str = Path(
        ...,
        min_length=1,
        max_length=50,
        description="模块 ID",
    ),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:permission:view")),
):
    """
    获取指定模块的所有权限。
    - 权限：`auth:permission:view`
    - 路径参数：`module_id`
    """
    data = await permission_service.get_permissions_by_module(db, module_id)
    return success_response(data=data)
