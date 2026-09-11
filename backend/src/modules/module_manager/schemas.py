"""模块管理模块 - Pydantic Schema。"""
from datetime import datetime
from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, ConfigDict, Field


class MenuItem(BaseModel):
    id: str
    parent_id: Optional[str] = None
    title: str
    icon: Optional[str] = None
    path: str
    component: str
    permission: Optional[str] = None
    order: int = 0


class ModuleInstallReq(BaseModel):
    """从受控目录安装模块（source_path 必须在 MODULE_UPLOAD_DIR 下）。"""
    install_type: Literal["path"] = Field("path", description="v1.1 起仅支持 path（ZIP 请使用 /upload）")
    source_path: str = Field(..., description="源码目录路径（必须位于 MODULE_UPLOAD_DIR 下）")


class ModuleUpgradeReq(BaseModel):
    """升级模块。"""
    install_type: Literal["zip", "path"] = Field(..., description="升级源类型")
    file_path: Optional[str] = Field(None, description="ZIP 包路径（zip 时必填）")
    source_path: Optional[str] = Field(None, description="源码目录路径（path 时必填）")


class ModuleConfigUpdate(BaseModel):
    config: Dict[str, Any] = Field(..., description="模块配置对象")


class ModuleOut(BaseModel):
    id: str
    name: str
    version: str
    description: Optional[str] = None
    author: Optional[str] = None
    homepage: Optional[str] = None
    status: str
    entry_backend: str
    entry_frontend: Optional[str] = None
    dependencies: List[str] = []
    menus: List[MenuItem] = []
    config: Dict[str, Any] = {}
    installed_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class ModuleUploadResp(BaseModel):
    """ZIP 上传安装响应。"""
    module_id: str
    version: str
    status: str
    saved_path: str
