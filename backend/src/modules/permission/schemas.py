from datetime import datetime

from pydantic import BaseModel


class PermissionOut(BaseModel):
    id: int
    code: str
    name: str
    module_id: str
    resource: str
    action: str
    created_at: datetime
