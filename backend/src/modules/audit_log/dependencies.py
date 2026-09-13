"""日志审计模块 - 依赖注入（复用认证模块）。"""
from src.modules.auth.dependencies import (
    CurrentUser,
    require_permission,
)
__all__ = ["CurrentUser", "require_permission"]
