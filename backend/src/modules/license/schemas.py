"""License 管理模块 - Pydantic Schema。
v1.1 变更：
- P1-7：LicenseStatusOut 移除 machine_code / current_machine_code，
        严格遵循《API接口文档》§ 8.1 契约。
- P2-3：删除未使用的 LicensePayload / LicenseFileContent。
- P2-7：LicenseActivateReq.activation_code 增加 max_length=255。
"""
from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, Field
class LicenseStatusOut(BaseModel):
    """
    License 状态响应（严格遵循《API接口文档》§ 8.1）。
    v1.1 起不再包含 machine_code / current_machine_code 字段。
    机器码信息仅用于服务端日志排障。
    """
    is_valid: bool = Field(..., description="当前 License 是否有效")
    license_type: Optional[str] = Field(None, description="授权类型")
    expires_at: Optional[datetime] = Field(None, description="过期时间")
    days_remaining: Optional[int] = Field(None, description="剩余天数")
    max_users: Optional[int] = Field(None, description="最大用户数（NULL 表示不限制）")
    current_users: int = Field(..., description="当前用户数")
    authorized_modules: List[str] = Field(
        default_factory=list, description="授权模块 ID 列表"
    )
    is_expired: bool = Field(False, description="是否已过期")
    is_expiring_soon: bool = Field(False, description="是否即将过期（≤7 天）")
class LicenseActivateReq(BaseModel):
    """在线激活请求。"""
    activation_code: str = Field(
        ...,
        min_length=1,
        max_length=255,
        description="激活码（1-255 位）",
    )
    machine_code: str = Field(
        ...,
        min_length=1,
        max_length=255,
        description="机器码（1-255 位）",
    )
class ModuleAuthorizationOut(BaseModel):
    """模块授权状态。"""
    module_id: str
    module_name: str
    is_authorized: bool
    expires_at: Optional[datetime] = None
