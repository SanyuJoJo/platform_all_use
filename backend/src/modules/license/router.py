from fastapi import APIRouter, Depends, File, Form, UploadFile
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .service import get_license_status, import_license, activate_license, get_module_auth_status
router = APIRouter(prefix="/api/v1/license", tags=["License"])
@router.get("/status")
async def license_status(
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("license:license:view"))
):
    status = get_license_status()
    return success_response(data=status)
@router.post("/import")
async def import_license_endpoint(
    license_file: UploadFile = File(None),
    activation_code: str = Form(None),
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("license:license:create"))
):
    if not license_file and not activation_code:
        raise PlatformException(code=90001, message="请提供 License 文件或激活码", status_code=400)
    file_content = await license_file.read() if license_file else None
    result = import_license(file_content, activation_code)
    return success_response(data=result, message="License导入成功")
@router.post("/activate")
async def activate_license_endpoint(
    req: dict,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("license:license:create"))
):
    activation_code = req.get("activation_code")
    machine_code = req.get("machine_code")
    if not activation_code or not machine_code:
        raise PlatformException(code=90001, message="激活码和机器码不能为空", status_code=400)
    result = activate_license(activation_code, machine_code)
    return success_response(data=result, message="激活成功")
@router.get("/modules")
async def get_license_modules(
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("license:license:view"))
):
    data = get_module_auth_status()
    return success_response(data=data)
