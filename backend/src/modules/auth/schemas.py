"""
认证模块 - Pydantic Schema
说明：密码长度等业务规则由 service 层校验，确保返回标准错误码
（10003 / 10004 / 10007），而不是 Pydantic 的 422 通用校验错误。
"""
from typing import List, Optional
from pydantic import BaseModel, Field
class LoginReq(BaseModel):
    """登录请求"""
    username: str = Field(..., min_length=1, max_length=50, description="用户名")
    password: str = Field(..., min_length=1, description="密码")
class RefreshReq(BaseModel):
    """刷新 Token 请求（兼容请求体方式）"""
    refresh_token: Optional[str] = Field(None, description="刷新令牌")
class ChangePasswordReq(BaseModel):
    """
    修改密码请求。
    - old_password：不做长度限制，由 service 校验后返回 10003；
    - new_password/confirm_password：长度 6-20 由 service 校验后返回 10007。
    """
    old_password: str = Field(..., min_length=0, description="旧密码")
    new_password: str = Field(..., min_length=0, description="新密码")
    confirm_password: str = Field(..., min_length=0, description="确认密码")
class UserInfo(BaseModel):
    """当前用户信息"""
    id: int
    username: str
    nickname: str
    email: Optional[str] = None
    avatar: Optional[str] = None
    status: int
    roles: List[str] = []
    permissions: List[str] = []
class TokenResp(BaseModel):
    """登录响应"""
    access_token: str
    refresh_token: str
    token_type: str = "bearer"
    expires_in: int
    user: UserInfo
class RefreshResp(BaseModel):
    """刷新响应（v1.2 起包含轮换后的 refresh_token）"""
    access_token: str
    refresh_token: str
    expires_in: int
