from fastapi import APIRouter, Depends, Query, Path
from typing import Optional
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .schemas import UserCreate, UserUpdate, UserStatusUpdate, UserPasswordReset
from .service import list_users, create_user, get_user_by_id, update_user, delete_user, patch_status, reset_password
router = APIRouter(prefix="/api/v1/auth/users", tags=["User"])
@router.get("/")
async def get_users(
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    keyword: Optional[str] = None,
    status: Optional[int] = None,
    role_id: Optional[int] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:view"))
):
    result = list_users(page, page_size, keyword, status, role_id)
    return success_response(data=result)
@router.post("/")
async def create_new_user(
    data: UserCreate,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:create"))
):
    user = create_user(data.dict())
    return success_response(data=user, message="用户创建成功")
@router.get("/{user_id}")
async def get_user_detail(
    user_id: int = Path(..., gt=0),
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:view"))
):
    user = get_user_by_id(user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    user["roles"] = [ROLES[rid] for rid in user["role_ids"] if rid in ROLES]  # 需导入ROLES
    return success_response(data=user)
@router.put("/{user_id}")
async def update_user_info(
    user_id: int,
    data: UserUpdate,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:edit"))
):
    updated = update_user(user_id, data.dict(exclude_unset=True))
    return success_response(data=updated, message="用户更新成功")
@router.delete("/{user_id}")
async def delete_user_endpoint(
    user_id: int,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:delete"))
):
    delete_user(user_id)
    return success_response(message="删除成功", data=None)
@router.patch("/{user_id}/status")
async def update_user_status(
    user_id: int,
    data: UserStatusUpdate,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:edit"))
):
    updated = patch_status(user_id, data.status)
    return success_response(data=updated, message="状态更新成功")
@router.patch("/{user_id}/password")
async def reset_user_password(
    user_id: int,
    data: UserPasswordReset,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:user:edit"))
):
    reset_password(user_id, data.new_password)
    return success_response(message="密码重置成功", data=None)
