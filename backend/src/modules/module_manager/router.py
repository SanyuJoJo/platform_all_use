"""模块管理模块 - 路由定义（v1.3）。"""
from typing import Optional

from fastapi import (
    APIRouter,
    Depends,
    File,
    Path,
    Query,
    Request,
    UploadFile,
)
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.database import get_db
from src.core.response import success_response
from src.modules.module_manager import service
from src.modules.module_manager.dependencies import CurrentUser, require_permission
from src.modules.module_manager.schemas import (
    ModuleConfigUpdate,
    ModuleInstallReq,
    ModuleUpgradeReq,
)

router = APIRouter(prefix="/api/v1/modules", tags=["Module Manager"])


@router.get("")
async def list_modules(
    request: Request,
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    status: Optional[str] = Query(None, description="状态：active/inactive"),
    keyword: Optional[str] = Query(None, description="搜索关键词"),
    filter_menus: bool = Query(True),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:view")),
):
    user_permissions = current_user.get("permissions", []) if filter_menus else None
    data = await service.list_modules(
        db, page, page_size, status, keyword, user_permissions=user_permissions
    )
    return success_response(data=data)


@router.post("")
async def install_module(
    req: ModuleInstallReq,
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:create")),
):
    data = await service.install_module(
        db,
        install_type="path",
        file_path=None,
        source_path=req.source_path,
        operator=current_user,
    )
    return success_response(data=data, message="模块安装成功")


@router.post("/upload")
async def upload_and_install(
    file: UploadFile = File(..., description="模块 ZIP 包"),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:create")),
):
    content = await file.read()
    saved_path = await service.save_uploaded_zip(file.filename or "module.zip", content)
    data = await service.install_module(
        db,
        install_type="zip",
        file_path=str(saved_path),
        source_path=None,
        operator=current_user,
    )
    return success_response(data=data, message="模块上传并安装成功")


@router.delete("/{module_id}")
async def uninstall_module(
    request: Request,
    module_id: str = Path(..., min_length=1, max_length=50),
    force: bool = Query(False),
    drop_tables: bool = Query(False),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:delete")),
):
    """卸载模块（v1.3：传入 app 以清理 _loaded_modules）。"""
    await service.uninstall_module(
        db,
        module_id,
        force=force,
        drop_tables=drop_tables,
        operator=current_user,
        app=request.app,
    )
    return success_response(message="卸载成功", data=None)


@router.post("/{module_id}/enable")
async def enable_module(
    request: Request,
    module_id: str = Path(..., min_length=1, max_length=50),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:edit")),
):
    data = await service.enable_module(db, module_id, current_user, request.app)
    return success_response(data=data, message="模块已启用")


@router.post("/{module_id}/disable")
async def disable_module(
    request: Request,
    module_id: str = Path(..., min_length=1, max_length=50),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:edit")),
):
    """停用模块（v1.3：传入 app 以清理 _loaded_modules）。"""
    data = await service.disable_module(db, module_id, current_user, request.app)
    return success_response(data=data, message="模块已停用")


@router.post("/{module_id}/upgrade")
async def upgrade_module(
    request: Request,
    req: ModuleUpgradeReq,
    module_id: str = Path(..., min_length=1, max_length=50),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:create")),
):
    """升级模块（v1.3：传入 app 以支持 active 模块重载路由）。"""
    data = await service.upgrade_module(
        db,
        module_id,
        install_type=req.install_type,
        file_path=req.file_path,
        source_path=req.source_path,
        operator=current_user,
        app=request.app,
    )
    return success_response(data=data, message="模块升级成功")


@router.get("/{module_id}/config")
async def get_module_config(
    module_id: str = Path(..., min_length=1, max_length=50),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("module_manager:module:view")),
):
    data = await service.get_module_config(db, module_id)
    return success_response(data=data)


@router.put("/{module_id}/config")
async def update_module_config(
    req: ModuleConfigUpdate,
    module_id: str = Path(..., min_length=1, max_length=50),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(require_permission("module_manager:module:edit")),
):
    data = await service.update_module_config(db, module_id, req.config, current_user)
    return success_response(data=data, message="配置更新成功")
