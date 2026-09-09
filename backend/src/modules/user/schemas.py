from pydantic import BaseModel, Field, validator
from typing import List, Optional
from datetime import datetime

class UserBase(BaseModel):
    username: str = Field(..., min_length=3, max_length=20, pattern=r"^[a-zA-Z0-9_]+$")  # 改这里
    nickname: str = Field(..., min_length=1, max_length=50)
    email: Optional[str] = None
    status: int = 1
    role_ids: List[int] = []

class UserCreate(UserBase):
    password: str = Field(..., min_length=6, max_length=20)

class UserUpdate(BaseModel):
    nickname: Optional[str] = Field(None, min_length=1, max_length=50)
    email: Optional[str] = None
    status: Optional[int] = None
    role_ids: Optional[List[int]] = None

class UserOut(BaseModel):
    id: int
    username: str
    nickname: str
    email: Optional[str] = None
    avatar: Optional[str] = None
    status: int
    roles: List[dict] = []
    created_at: datetime
    updated_at: datetime

class UserStatusUpdate(BaseModel):
    status: int = Field(..., ge=0, le=1)

class UserPasswordReset(BaseModel):
    new_password: str = Field(..., min_length=6, max_length=20)
