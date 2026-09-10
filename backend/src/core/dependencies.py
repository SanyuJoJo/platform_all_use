"""
核心依赖兼容层。
实际实现位于 src.modules.auth.dependencies。
保留 oauth2_scheme 兼容别名，指向认证模块的 bearer_scheme，
避免旧引用（如 Swagger 定义）因移除导出而报错。
"""
from src.modules.auth.dependencies import (
    CurrentUser,
    bearer_scheme,
    get_current_user,
    require_permission,
)
# 兼容别名：历史上使用 OAuth2PasswordBearer；现统一使用 HTTPBearer
oauth2_scheme = bearer_scheme
__all__ = [
    "CurrentUser",
    "bearer_scheme",
    "get_current_user",
    "oauth2_scheme",
    "require_permission",
]
