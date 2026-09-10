"""
角色管理模块 - Pydantic Schema（v1.3）
v1.3 变更：
- P2-NEW-B：RoleUpdate 显式声明 `model_config = ConfigDict(extra="forbid")`，
        客户端传 `code` / `is_system` / 任何未定义字段返回 422 / 90004。
        这是 v1.1 评审 P2-NEW-6「Schema 白名单之外的第二道防线」的真正落地
        实现（v1.2 的路由防御因 Pydantic 默认 extra="ignore" 实际为死代码）。
v1.1 变更：
- P0-5：移除所有 min_length / max_length，长度与格式校验统一由 Service
        层执行并返回标准业务错误码（90001 / 20001 / 20004 / 20054）。
- P2-1：RoleCreate.code 不再使用 Pydantic pattern 限制；改由 Service
        层按 ^[a-z][a-z0-9_]*$ 校验。
- P2-2：RoleOut 显式包含 is_system 字段。
设计说明：
- RoleCreate / RoleUpdate **不含** is_system 字段，API 无法创建或修改
  系统内置角色标志；系统内置角色仅由认证模块 ensure_auth_seed_data 创建。
- code 字段在 RoleUpdate 中**不存在**，角色编码创建后不可修改；
  客户端若传 `code`，由 `extra="forbid"` 返回 422 / 90004。
"""
from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel, ConfigDict, Field
class RoleCreate(BaseModel):
    """创建角色请求。
    注意：
        name / code 的长度与格式校验全部交由 Service 层完成，以便统一返回
        业务错误码（90001），而不是 Pydantic 的 422/90004。
    """
    name: str = Field(..., description="角色名称（长度 1-50 由 service 校验）")
    code: str = Field(
        ...,
        description="角色编码（长度 1-50，^[a-z][a-z0-9_]*$ 由 service 校验）",
    )
    description: Optional[str] = Field(None, description="角色描述（长度 ≤255 由 service 校验）")
    permission_codes: List[str] = Field(
        default_factory=list,
        description="权限编码列表，格式 {module}:{resource}:{action}",
    )
class RoleUpdate(BaseModel):
    """更新角色请求（所有字段可选）。
    v1.3 新增 P2-NEW-B：
        - `model_config = ConfigDict(extra="forbid")` 显式拒绝未知字段；
        - 客户端传 `code` / `is_system` / 任何未定义字段 → 422 / 90004；
        - 这是「Schema 白名单之外的第二道防线」的正式落地实现。
    注意：
        - code 字段**不在本 Schema 中**，角色编码创建后不可修改；
        - is_system 字段**不在本 Schema 中**，系统内置标志不可通过 API 修改。
    """
    model_config = ConfigDict(extra="forbid")
    name: Optional[str] = Field(None, description="角色名称（长度 1-50 由 service 校验）")
    description: Optional[str] = Field(None, description="角色描述（长度 ≤255 由 service 校验）")
    permission_codes: Optional[List[str]] = Field(
        None,
        description="权限编码列表（全量覆盖，格式 {module}:{resource}:{action}）",
    )
class RoleOut(BaseModel):
    """角色响应对象。
    字段说明：
        - id: 角色 ID
        - name: 角色名称
        - code: 角色编码（创建后不可变）
        - description: 角色描述
        - is_system: 1-系统内置（不可删除）0-自定义
        - permission_codes: 权限编码列表（已排序，去重）
        - created_at / updated_at: 时间戳
    """
    id: int
    name: str
    code: str
    description: Optional[str] = None
    is_system: int
    permission_codes: List[str] = []
    created_at: datetime
    updated_at: datetime
