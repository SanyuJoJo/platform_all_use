from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
class MenuItem(BaseModel):
    id: str
    parent_id: Optional[str] = None
    title: str
    icon: Optional[str] = None
    path: str
    component: str
    permission: Optional[str] = None
    order: int = 0
class ModuleOut(BaseModel):
    id: str
    name: str
    version: str
    description: str
    author: Optional[str] = None
    status: str
    entry_frontend: Optional[str] = None
    dependencies: List[str] = []
    installed_at: datetime
    updated_at: datetime
    menus: List[MenuItem] = []
