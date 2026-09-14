"""用户管理模块 - 路由定义（v1.1）。

v1.1 修复：
- F821：新增模块级 ROLES 常量，供 get_user_detail 返回角色详情使用。
- I001：整理导入顺序。
- F401：移除未使用的 Query 导入。
"""
from typing import Optional

from fastapi import APIRouter, Depends, Path, Query

from src.core.dependencies import get_current_user, require_permission
from src.core.exceptions import PlatformException
from src.core.response import success_response
from src.modules.user.schemas import (
    UserCreate,
    UserPasswordReset,
    UserStatusUpdate,
    UserUpdate,
)
from src.modules.user.service import (
    create_user,
    delete_user,
    get_user_by_id,
    list_users,
    patch_status,
    reset_password,
    update_user,
)

router = APIRouter(prefix="/api/v1/auth/users", tags=["User"])


# ---------------------------------------------------------------------------
# 角色字典（与 src/modules/role/service.py 的 MOCK_ROLES 保持一致）
# 用于在 get_user_detail 中填充 user["roles"] 字段。
# ---------------------------------------------------------------------------
ROLES: dict[int, dict] = {
    1: {"id": 1, "name": "管理员", "code": "admin"},
    2: {"id": 2, "name": "访客", "code": "guest"},
}


@router.get("")
async def list_users_endpoint(
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    keyword: Optional[str] = None,
    status: Optional[int] = None,
    role_id: Optional[int] = None,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:view")),
):
    result = list_users(page, page_size, keyword, status, role_id)
    return success_response(data=result)


@router.post("")
async def create_new_user(
    data: UserCreate,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:create")),
):
    user = create_user(data.model_dump())
    return success_response(data=user)


@router.get("/{user_id}")
async def get_user_detail(
    user_id: int = Path(..., gt=0),
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:view")),
):
    user = get_user_by_id(user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    user["roles"] = [
        ROLES[rid] for rid in user.get("role_ids", []) if rid in ROLES
    ]
    return success_response(data=user)


@router.put("/{user_id}")
async def update_user_endpoint(
    user_id: int,
    data: UserUpdate,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:edit")),
):
    updated = update_user(user_id, data.model_dump(exclude_unset=True))
    return success_response(data=updated)


@router.delete("/{user_id}")
async def delete_user_endpoint(
    user_id: int,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:delete")),
):
    delete_user(user_id)
    return success_response(message="删除成功", data=None)


@router.patch("/{user_id}/status")
async def patch_user_status(
    user_id: int,
    data: UserStatusUpdate,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:edit")),
):
    updated = patch_status(user_id, data.status)
    return success_response(data=updated)


@router.patch("/{user_id}/password")
async def reset_user_password(
    user_id: int,
    data: UserPasswordReset,
    current_user: dict = Depends(get_current_user),
    _=Depends(require_permission("auth:user:edit")),
):
    reset_password(user_id, data.new_password)
    return success_response(message="密码重置成功", data=None)
