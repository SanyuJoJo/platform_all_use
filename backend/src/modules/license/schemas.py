from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime
class LicenseStatus(BaseModel):
    is_valid: bool
    license_type: str
    expires_at: Optional[datetime]
    days_remaining: Optional[int]
    max_users: Optional[int]
    current_users: Optional[int]
    authorized_modules: List[str]
    is_expired: bool
    is_expiring_soon: bool
class ModuleAuthStatus(BaseModel):
    module_id: str
    module_name: str
    is_authorized: bool
    expires_at: Optional[datetime]
