from pydantic import BaseModel
from typing import Optional
from datetime import datetime
class AuditLogOut(BaseModel):
    id: int
    user_id: Optional[int]
    username: Optional[str]
    module_id: str
    action: str
    resource: Optional[str]
    resource_id: Optional[str]
    detail: Optional[str]
    ip: Optional[str]
    user_agent: Optional[str]
    status: str
    error_code: Optional[int]
    request_id: Optional[str]
    created_at: datetime
