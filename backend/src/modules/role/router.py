from fastapi import APIRouter, Depends, Query, Path
from typing import Optional
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .schemas import RoleCreate, RoleUpdate
from .service import list_roles, create_role, get_role_by_id, update_role, delete_role
router = APIRouter(prefix="/api/v1/auth/roles", tags=["Role"])
@router.get("/")
async def get_roles(
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    keyword: Optional[str] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:role:view"))
):
    result = list_roles(page, page_size, keyword)
    return success_response(data=result)
@router.post("/")
async def create_new_role(
    data: RoleCreate,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:role:create"))
):
    role = create_role(data.dict())
    return success_response(data=role, message="角色创建成功")
@router.get("/{role_id}")
async def get_role_detail(
    role_id: int = Path(..., gt=0),
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:role:view"))
):
    role = get_role_by_id(role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)
    return success_response(data=role)
@router.put("/{role_id}")
async def update_role_info(
    role_id: int,
    data: RoleUpdate,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:role:edit"))
):
    updated = update_role(role_id, data.dict(exclude_unset=True))
    return success_response(data=updated, message="角色更新成功")
@router.delete("/{role_id}")
async def delete_role_endpoint(
    role_id: int,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:role:delete"))
):
    delete_role(role_id)
    return success_response(message="删除成功", data=None)
