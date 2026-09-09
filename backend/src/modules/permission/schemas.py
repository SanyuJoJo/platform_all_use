from pydantic import BaseModel
from typing import Optional
from datetime import datetime
class PermissionOut(BaseModel):
    id: int
    code: str
    name: str
    module_id: str
    resource: str
    action: str
    created_at: datetime
