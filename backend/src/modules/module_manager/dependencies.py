"""模块管理模块 - 依赖注入。"""
from src.modules.auth.dependencies import CurrentUser, require_permission

__all__ = ["CurrentUser", "require_permission"]
