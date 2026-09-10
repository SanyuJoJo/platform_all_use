"""
用户管理模块 - 路由定义
路由前缀：/api/v1/auth/users
权限点：auth:user:view / auth:user:create / auth:user:edit / auth:user:delete
"""
from typing import Optional
from fastapi import APIRouter, Depends, Path, Query
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.database import get_db
from src.core.response import success_response
from src.modules.auth import user_service
from src.modules.auth.dependencies import (
    CurrentUser,
    require_permission,
)
from src.modules.auth.user_schemas import (
    UserCreate,
    UserPasswordReset,
    UserStatusUpdate,
    UserUpdate,
)
router = APIRouter(prefix="/api/v1/auth/users", tags=["User"])
@router.get("")
async def list_users(
    page: int = Query(1, ge=1, description="页码，默认 1"),
    page_size: int = Query(20, ge=1, le=100, description="每页条数，默认 20，最大 100"),
    keyword: Optional[str] = Query(None, description="关键字：用户名/昵称/邮箱"),
    status: Optional[int] = Query(None, ge=0, le=1, description="状态：1-启用 0-禁用"),
    role_id: Optional[int] = Query(None, description="按角色 ID 筛选"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:user:view")),
):
    """查询用户列表（分页、关键字、状态、角色筛选）。"""
    data = await user_service.list_users(
        db, page=page, page_size=page_size, keyword=keyword, status=status, role_id=role_id
    )
    return success_response(data=data)
@router.post("")
async def create_user(
    req: UserCreate,
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:user:create")),
):
    """创建用户。"""
    data = await user_service.create_user(db, req.model_dump(), current_user)
    return success_response(data=data, message="用户创建成功")
@router.get("/{user_id}")
async def get_user_detail(
    user_id: int = Path(..., gt=0, description="用户 ID"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("auth:user:view")),
):
    """获取用户详情。"""
    data = await user_service.get_user_detail(db, user_id)
    return success_response(data=data)
@router.put("/{user_id}")
async def update_user(
    req: UserUpdate,
    user_id: int = Path(..., gt=0, description="用户 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:user:edit")),
):
    """更新用户信息（role_ids 全量覆盖）。"""
    payload = req.model_dump(exclude_unset=True)
    data = await user_service.update_user(db, user_id, payload, current_user)
    return success_response(data=data, message="用户更新成功")
@router.delete("/{user_id}")
async def delete_user(
    user_id: int = Path(..., gt=0, description="用户 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:user:delete")),
):
    """删除用户。"""
    await user_service.delete_user(db, user_id, current_user)
    return success_response(message="删除成功", data=None)
@router.patch("/{user_id}/status")
async def update_user_status(
    req: UserStatusUpdate,
    user_id: int = Path(..., gt=0, description="用户 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:user:edit")),
):
    """启用/禁用用户。"""
    data = await user_service.update_user_status(db, user_id, req.status, current_user)
    return success_response(data=data, message="状态更新成功")
@router.patch("/{user_id}/password")
async def reset_user_password(
    req: UserPasswordReset,
    user_id: int = Path(..., gt=0, description="用户 ID"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("auth:user:edit")),
):
    """管理员重置用户密码。"""
    await user_service.reset_user_password(
        db, user_id, req.new_password, current_user
    )
    return success_response(message="密码重置成功", data=None)
