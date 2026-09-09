from datetime import datetime, timezone, timedelta
from typing import List, Dict
from src.core.exceptions import PlatformException
# 模拟 License 数据
MOCK_LICENSE = {
    "is_valid": True,
    "license_type": "enterprise",
    "expires_at": datetime(2027, 8, 28, 23, 59, 59, tzinfo=timezone.utc),
    "max_users": 100,
    "current_users": 45,
    "authorized_modules": ["auth", "audit_log", "customer_relation", "project_management"],
    "is_expired": False,
    "is_expiring_soon": False,
}
# 模块授权状态（基于授权模块列表）
MODULE_AUTH_STATUS = [
    {"module_id": "customer_relation", "module_name": "客户关系管理", "is_authorized": True, "expires_at": MOCK_LICENSE["expires_at"]},
    {"module_id": "project_management", "module_name": "项目管理", "is_authorized": True, "expires_at": MOCK_LICENSE["expires_at"]},
    {"module_id": "audit_log", "module_name": "日志审计", "is_authorized": True, "expires_at": MOCK_LICENSE["expires_at"]},
    {"module_id": "module_manager", "module_name": "模块管理", "is_authorized": False, "expires_at": None},  # 未授权
]
def get_license_status() -> dict:
    return MOCK_LICENSE
def import_license(file_content: bytes = None, activation_code: str = None) -> dict:
    # 模拟导入，简单更新状态
    MOCK_LICENSE["is_valid"] = True
    MOCK_LICENSE["expires_at"] = datetime(2028, 1, 1, 23, 59, 59, tzinfo=timezone.utc)
    MOCK_LICENSE["authorized_modules"] = ["auth", "audit_log", "customer_relation", "project_management", "module_manager"]
    # 更新模块授权
    for item in MODULE_AUTH_STATUS:
        if item["module_id"] in MOCK_LICENSE["authorized_modules"]:
            item["is_authorized"] = True
            item["expires_at"] = MOCK_LICENSE["expires_at"]
        else:
            item["is_authorized"] = False
            item["expires_at"] = None
    return {
        "license_type": "enterprise",
        "expires_at": MOCK_LICENSE["expires_at"],
        "authorized_modules": MOCK_LICENSE["authorized_modules"],
    }
def activate_license(activation_code: str, machine_code: str) -> dict:
    # 模拟激活
    if activation_code != "ABCD-1234-EFGH-5678":
        raise PlatformException(code=50005, message="激活码无效", status_code=400)
    # 模拟成功
    MOCK_LICENSE["is_valid"] = True
    MOCK_LICENSE["expires_at"] = datetime(2028, 1, 1, 23, 59, 59, tzinfo=timezone.utc)
    return {"message": "激活成功"}
def get_module_auth_status() -> List[dict]:
    return MODULE_AUTH_STATUS
