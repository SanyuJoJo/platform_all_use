"""License 管理模块 - 常量定义。
v1.1 变更：
- P0-5：新增 LICENSE_KEY_MIN_LEN / LICENSE_KEY_MAX_LEN，
        供 _persist_license 中的 license_key 长度校验使用。
"""
from typing import Final

# License 类型白名单（与数据库设计文档 § 6.1 一致）
LICENSE_TYPES: Final = ("trial", "standard", "enterprise")
# License 文件允许的扩展名
LICENSE_FILE_EXTENSIONS: Final = (".lic", ".json")
# License 文件最大大小（1MB，防止内存爆炸）
LICENSE_FILE_MAX_SIZE: Final = 1 * 1024 * 1024
# 过期提醒阈值（天）
EXPIRING_SOON_DAYS: Final = 7
# 受 License 限制的核心模块白名单（始终允许访问）
# 避免平台自身被锁死
CORE_MODULES_BYPASS: Final = frozenset(
    {"auth", "platform", "module_manager", "audit_log", "license"}
)
# 签名算法（预留 RSA 升级点）
SIGNATURE_ALGORITHM: Final = "HMAC-SHA256"
# License 文件必需字段
REQUIRED_PAYLOAD_FIELDS: Final = (
    "license_key",
    "license_type",
    "issued_at",
    "expires_at",
)
# license_key 长度约束（v1.1 P0-5 新增）
LICENSE_KEY_MIN_LEN: Final = 1
LICENSE_KEY_MAX_LEN: Final = 255
# 部分唯一索引名（v1.2 R-3 新增：保证同一时刻仅一个 is_active=1）
ACTIVE_LICENSE_UNIQUE_INDEX: Final = "uq_license_license_active"
