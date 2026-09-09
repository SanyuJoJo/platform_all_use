from fastapi import APIRouter, Depends, Query, Path
from typing import Optional
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .service import (
    get_modules,
    install_module,
    uninstall_module,
    enable_module,
    disable_module,
    get_module_config,
    update_module_config,
)

router = APIRouter(prefix="/api/v1/modules", tags=["Module Manager"])


@router.get("/")
async def list_modules(
    status: Optional[str] = Query(None, regex="^(active|inactive)$"),
    keyword: Optional[str] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:view"))
):
    """查询模块列表（含菜单权限过滤）"""
    user_permissions = current_user.get("permissions", [])
    modules_data = get_modules(status, keyword, user_permissions)
    total = len(modules_data)
    return success_response(
        data={
            "items": modules_data,
            "total": total,
            "page": 1,
            "page_size": total if total > 0 else 1,
            "pages": 1
        }
    )


@router.post("/")
async def install_module_endpoint(
    req: dict,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:create"))
):
    """安装模块（支持 zip 或 path 方式）"""
    install_type = req.get("install_type")
    if install_type == "zip":
        file_path = req.get("file_path")
        if not file_path:
            raise PlatformException(code=90001, message="缺少 file_path", status_code=400)
        source_path = None
    elif install_type == "path":
        source_path = req.get("source_path")
        if not source_path:
            raise PlatformException(code=90001, message="缺少 source_path", status_code=400)
        file_path = None
    else:
        raise PlatformException(code=90001, message="install_type 必须为 zip 或 path", status_code=400)

    module = install_module(install_type, file_path, source_path)
    return success_response(data=module, message="模块安装成功")


@router.delete("/{module_id}")
async def uninstall_module_endpoint(
    module_id: str,
    force: bool = Query(False),
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:delete"))
):
    """卸载模块（支持强制）"""
    uninstall_module(module_id, force)
    return success_response(message="卸载成功", data=None)


@router.post("/{module_id}/enable")
async def enable_module_endpoint(
    module_id: str,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:edit"))
):
    """启用模块"""
    module = enable_module(module_id)
    return success_response(data=module, message="模块已启用")


@router.post("/{module_id}/disable")
async def disable_module_endpoint(
    module_id: str,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:edit"))
):
    """停用模块"""
    module = disable_module(module_id)
    return success_response(data=module, message="模块已停用")


@router.get("/{module_id}/config")
async def get_module_config_endpoint(
    module_id: str,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:view"))
):
    """获取模块配置"""
    config = get_module_config(module_id)
    return success_response(data=config)


@router.put("/{module_id}/config")
async def update_module_config_endpoint(
    module_id: str,
    req: dict,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("module_manager:module:edit"))
):
    """更新模块配置（全量覆盖）"""
    config = req.get("config")
    if config is None:
        raise PlatformException(code=90001, message="缺少 config 字段", status_code=400)
    updated = update_module_config(module_id, config)
    return success_response(data=updated, message="配置更新成功")
