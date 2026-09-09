from fastapi import APIRouter, Depends, Query
from typing import Optional
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from .service import list_permissions, get_permissions_by_module
router = APIRouter(prefix="/api/v1/auth/permissions", tags=["Permission"])
@router.get("/")
async def get_permissions(
    module_id: Optional[str] = None,
    resource: Optional[str] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:permission:view"))
):
    perms = list_permissions(module_id, resource)
    return success_response(data=perms)
@router.get("/modules/{module_id}")
async def get_module_permissions(
    module_id: str,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("auth:permission:view"))
):
    perms = get_permissions_by_module(module_id)
    return success_response(data=perms)
