from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime

class RoleBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=50)
    code: str = Field(..., min_length=1, max_length=50, pattern=r"^[a-z_]+$")  # 改这里
    description: Optional[str] = None

class RoleCreate(RoleBase):
    permission_codes: List[str] = []

class RoleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    permission_codes: Optional[List[str]] = None

class RoleOut(RoleBase):
    id: int
    is_system: bool
    permission_codes: List[str]
    created_at: datetime
    updated_at: datetime
