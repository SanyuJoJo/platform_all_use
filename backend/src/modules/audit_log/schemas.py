"""日志审计模块 - Pydantic Schema。"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, ConfigDict
class AuditLogOut(BaseModel):
    """操作日志响应对象。"""
    id: int
    user_id: Optional[int] = None
    username: Optional[str] = None
    module_id: str
    action: str
    resource: Optional[str] = None
    resource_id: Optional[str] = None
    detail: Optional[str] = None
    ip: Optional[str] = None
    user_agent: Optional[str] = None
    status: str
    error_code: Optional[int] = None
    request_id: Optional[str] = None
    created_at: datetime
    model_config = ConfigDict(from_attributes=True)
