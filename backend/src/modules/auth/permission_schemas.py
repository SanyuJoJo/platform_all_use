"""
权限管理模块 - Pydantic Schema（v1.1，v1.2 未修改）
v1.1 变更：
- P1-4：删除未使用的 PermissionRegisterItem，减少维护成本。
- 保留 PermissionOut 用于权限查询接口响应。
"""
from datetime import datetime
from pydantic import BaseModel
class PermissionOut(BaseModel):
    """权限响应对象。"""
    id: int
    code: str
    name: str
    module_id: str
    resource: str
    action: str
    created_at: datetime
