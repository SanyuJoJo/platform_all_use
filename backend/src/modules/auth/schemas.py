from pydantic import BaseModel
from typing import List, Optional
class LoginReq(BaseModel):
    username: str
    password: str
class UserInfo(BaseModel):
    id: int
    username: str
    nickname: str
    email: Optional[str] = None
    avatar: Optional[str] = None
    status: int
    roles: List[str] = []
    permissions: List[str] = []
class TokenResp(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"
    expires_in: int = 1440
    user: UserInfo
