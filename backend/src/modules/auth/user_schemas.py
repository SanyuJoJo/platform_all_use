
"""
用户管理模块 - Pydantic Schema（v1.2）
v1.2 变更：
- P2-4：移除 nickname 的 min_length / max_length，改由 Service 层统一
        校验并返回 90001（请求参数错误），与错误码体系保持一致。
v1.1 变更：
- P1-4：UserCreate.username 去除 min_length / max_length 限制
- P2-5：UserOut.created_at / updated_at 改为必填
"""
from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, Field
class RoleBrief(BaseModel):
    """角色简要信息（用于用户对象内嵌）。"""
    id: int
    name: str
    code: str
class UserCreate(BaseModel):
    """
    创建用户请求。
    注意：
        username / password / nickname 的长度校验全部交由 Service 层完成，
        以便统一返回业务错误码，而不是 Pydantic 的 422/90004。
    """
    username: str = Field(..., description="用户名（长度 3-20 由 service 校验）")
    password: str = Field(..., description="密码（长度 6-20 由 service 校验）")
    nickname: str = Field(..., description="昵称（长度 1-50 由 service 校验）")
    email: Optional[str] = Field(None, max_length=100, description="邮箱")
    role_ids: List[int] = Field(default_factory=list, description="角色 ID 列表")
    status: int = Field(1, ge=0, le=1, description="状态：1-启用 0-禁用")
class UserUpdate(BaseModel):
    """更新用户请求（所有字段可选）。"""
    nickname: Optional[str] = Field(None, description="昵称（长度 1-50 由 service 校验）")
    email: Optional[str] = Field(None, max_length=100, description="邮箱")
    role_ids: Optional[List[int]] = Field(None, description="角色 ID 列表（全量覆盖）")
    status: Optional[int] = Field(None, ge=0, le=1, description="状态：1-启用 0-禁用")
class UserStatusUpdate(BaseModel):
    """启用/禁用请求。"""
    status: int = Field(..., ge=0, le=1, description="状态：1-启用 0-禁用")
class UserPasswordReset(BaseModel):
    """管理员重置用户密码请求。"""
    new_password: str = Field(..., description="新密码（长度 6-20 由 service 校验）")
class UserOut(BaseModel):
    """用户响应对象。"""
    id: int
    username: str
    nickname: str
    email: Optional[str] = None
    avatar: Optional[str] = None
    status: int
    roles: List[RoleBrief] = []
    created_at: datetime
    updated_at: datetime
