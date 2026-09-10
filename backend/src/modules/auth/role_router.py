"""
角色管理模块 - 路由定义（v1.3）
路由前缀：/api/v1/auth/roles
权限点：auth:role:view / auth:role:create / auth:role:edit / auth:role:delete
v1.3 变更：
- P2-NEW-B：删除 update_role 中的死代码防御（v1.2 的 `if "code" in payload
        or "is_system" in payload` 因 Pydantic 默认 extra="ignore" 永不触发）。
        改由 RoleUpdate 的 `model_config = ConfigDict(extra="forbid")` 在
        Pydantic 构造阶段直接返回 422 / 90004。
v1.1 变更：
- P1-1：路由路径由 "/" 改为 ""，与 API 文档 § 4.1 一致，避免 307 重定向。
- P1-5：将 `_ = Depends(require_permission(...))` 改为
        `current_user: CurrentUser = Depends(require_permission(...))`，
        并将 current_user 注入 Service 用于审计日志。
"""
from typing import Optional
from fastapi import APIRouter, Depends, Path, Query
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.database import get_db
from src.core.response import success_response
from src.modules.auth import role_service
from src.modules.auth.dependencies import CurrentUser, require_permission
from src.modules.auth.role_schemas import RoleCreate, RoleUpdate
router = APIRouter(prefix="/api/v1/auth/roles", tags=["Role"])
@router.get("")
async def list_roles(
    page: int = Query(1, ge=1, description="页码，默认 1"),
    page_size: int = Query(20, ge=1, le=100, description="每页条数，默认 20，最大 100"),
    keyword: Optional[str] = Query(None, description="关键字：角色名称/编码"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:role:view")),
):
    """查询角色列表（分页、关键字筛选）。
    - 权限：`auth:role:view`
    - 分页格式遵循 API 文档 § 1.3
    """
    data = await role_service.list_roles(
        db,
        page=page,
        page_size=page_size,
        keyword=keyword,
    )
    return success_response(data=data)
@router.post("")
async def create_role(
    req: RoleCreate,
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:role:create")),
):
    """创建角色。
    - 权限：`auth:role:create`
    - 请求体：`RoleCreate`（name / code / description / permission_codes）
    - 响应：新创建的角色对象
    """
    data = await role_service.create_role(db, req.model_dump(), current_user)
    return success_response(data=data, message="角色创建成功")
@router.get("/{role_id}")
async def get_role_detail(
    role_id: int = Path(..., gt=0, description="角色 ID"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:role:view")),
):
    """获取角色详情。
    - 权限：`auth:role:view`
    - 错误：角色不存在返回 20002
    """
    data = await role_service.get_role_detail(db, role_id)
    return success_response(data=data)
@router.put("/{role_id}")
async def update_role(
    req: RoleUpdate,
    role_id: int = Path(..., gt=0, description="角色 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:role:edit")),
):
    """更新角色信息。
    - 权限：`auth:role:edit`
    - `permission_codes` 若提供，执行**全量覆盖**（先删后插）
    - `code` 字段不可修改；`is_system` 字段不可修改
    v1.3 变更 P2-NEW-B：
        - 删除 v1.2 的路由死代码防御；
        - 改由 `RoleUpdate.model_config = ConfigDict(extra="forbid")` 在
          Pydantic 构造阶段拒绝未知字段，返回 422 / 90004。
    """
    # RoleUpdate 已启用 extra="forbid"，客户端传 code / is_system /
    # 任何未定义字段将在 Pydantic 构造阶段抛出 ValidationError，
    # 由全局异常处理器返回 422 / 90004。
    payload = req.model_dump(exclude_unset=True)
    data = await role_service.update_role(db, role_id, payload, current_user)
    return success_response(data=data, message="角色更新成功")
@router.delete("/{role_id}")
async def delete_role(
    role_id: int = Path(..., gt=0, description="角色 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:role:delete")),
):
    """删除角色。
    - 权限：`auth:role:delete`
    - 系统内置角色不可删除（20003）
    - 角色被用户使用时不可删除（20005）
    """
    await role_service.delete_role(db, role_id, current_user)
    return success_response(message="删除成功", data=None)
