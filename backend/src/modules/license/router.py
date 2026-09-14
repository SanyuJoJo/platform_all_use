"""License 管理模块 - 路由定义。
v1.1 变更：
- P2-7：import_license 的 activation_code 增加 max_length=255。
路由前缀：/api/v1/license
权限点：
- license:license:view   → /status, /modules
- license:license:create → /import, /activate
接口契约严格遵循《API接口文档》§ 八。
"""
import logging
from typing import Optional

from fastapi import APIRouter, Depends, File, Form, UploadFile
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.database import get_db
from src.core.exceptions import PlatformException
from src.core.response import success_response
from src.modules.license import service
from src.modules.license.constants import (
    LICENSE_FILE_EXTENSIONS,
    LICENSE_FILE_MAX_SIZE,
)
from src.modules.license.dependencies import CurrentUser, require_permission
from src.modules.license.schemas import LicenseActivateReq

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/v1/license", tags=["License"])
@router.get("/status")
async def get_license_status(
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("license:license:view")),
):
    """
    获取 License 状态。
    - 权限：`license:license:view`
    - License 不存在返回 50009
    - 返回 `is_valid` / `is_expired` / `is_expiring_soon` / `days_remaining`
      供前端做过期提醒
    """
    data = await service.get_license_status(db)
    return success_response(data=data)
@router.get("/modules")
async def get_module_authorization(
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("license:license:view")),
):
    """
    获取所有已安装模块的授权状态。
    - 权限：`license:license:view`
    - 返回每个模块是否在 License 授权列表中
    - 核心模块始终 `is_authorized=true`
    """
    data = await service.get_module_authorization(db)
    return success_response(data=data)
@router.post("/import")
async def import_license(
    license_file: Optional[UploadFile] = File(
        None, description="License 文件（.lic 或 .json），与 activation_code 二选一"
    ),
    activation_code: Optional[str] = Form(
        None,
        max_length=255,
        description="在线激活码（1-255 位），与 license_file 二选一",
    ),
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(
        require_permission("license:license:create")
    ),
):
    """
    导入 License。
    - 权限：`license:license:create`
    - 请求：multipart/form-data
    - `license_file`：.lic / .json 文件
    - `activation_code`：在线激活码（1-255 位，调用外部服务拉取 License）
    - 两者二选一，同时为空返回 50001
    - 签名验证失败返回 50003
    - License 已过期返回 50002
    - 重复导入返回 50004
    """
    file_content: Optional[bytes] = None
    if license_file is not None:
        filename = license_file.filename or ""
        lower = filename.lower()
        if not any(lower.endswith(ext) for ext in LICENSE_FILE_EXTENSIONS):
            raise PlatformException(
                code=50001,
                message=f"License 文件扩展名无效，仅支持 {LICENSE_FILE_EXTENSIONS}",
                status_code=400,
            )
        content = await license_file.read()
        if len(content) == 0:
            raise PlatformException(
                code=50001, message="License 文件为空", status_code=400
            )
        if len(content) > LICENSE_FILE_MAX_SIZE:
            raise PlatformException(
                code=50001,
                message=f"License 文件超过大小上限 {LICENSE_FILE_MAX_SIZE} 字节",
                status_code=400,
            )
        file_content = content
    data = await service.import_license(
        db,
        file_content=file_content,
        activation_code=activation_code,
        operator=current_user,
    )
    return success_response(data=data, message="License 导入成功")
@router.post("/activate")
async def activate_license(
    req: LicenseActivateReq,
    db: AsyncSession = Depends(get_db),
    current_user: CurrentUser = Depends(
        require_permission("license:license:create")
    ),
):
    """
    在线激活。
    - 权限：`license:license:create`
    - 请求体：`{activation_code, machine_code}`
    - 调用外部激活服务（`LICENSE_ACTIVATION_URL`）
    - v1.1：请求 machine_code 必须与当前机器码一致，否则返回 50006
    - 未配置激活服务返回 50005
    - 激活成功后按导入流程落库
    """
    data = await service.activate_license(
        db,
        activation_code=req.activation_code,
        machine_code=req.machine_code,
        operator=current_user,
    )
    return success_response(data=data, message="License 激活成功")
