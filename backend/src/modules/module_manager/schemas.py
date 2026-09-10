"""模块管理模块 - Pydantic Schema（mock 版）。"""

from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel


class MenuItem(BaseModel):
    """菜单项。"""

    id: str
    parent_id: Optional[str] = None
    title: str
    icon: Optional[str] = None
    path: str
    component: str
    permission: Optional[str] = None
    order: int = 0


class ModuleOut(BaseModel):
    """模块响应对象。"""

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
