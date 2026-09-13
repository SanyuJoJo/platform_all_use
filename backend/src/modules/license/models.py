"""License 管理模块 - ORM 模型。
表结构依据《数据库设计文档》§ 6.1。
v1.2 变更（R-3）：
- 增加部分唯一索引 uq_license_license_active（WHERE is_active = 1），
  从数据库层保证同一时刻最多一条 active 记录，消除并发导入
  不同 license_key 时产生多条 active 的边界问题。
"""
from datetime import datetime, timezone
from sqlalchemy import (
    Column,
    DateTime,
    Index,
    Integer,
    JSON,
    SmallInteger,
    String,
    text,
)
from sqlalchemy.sql import func
from src.core.database import Base
from src.modules.license.constants import ACTIVE_LICENSE_UNIQUE_INDEX
def _utcnow_naive() -> datetime:
    """UTC 当前时间（naive），与 SQLite DateTime 存储格式一致。"""
    return datetime.now(timezone.utc).replace(tzinfo=None)
class License(Base):
    """
    License 表（license_license）。
    字段说明详见《数据库设计文档》§ 6.1。
    - is_active：1-激活 0-未激活（导入后即为 1）
    - activated_at：首次导入/激活时间
    - machine_code：绑定的机器码，NULL 表示不绑定
    - authorized_modules：授权模块 ID 列表（JSON 数组）
    v1.2（R-3）：部分唯一索引保证同一时刻最多一条 is_active=1 记录。
    """
    __tablename__ = "license_license"
    id = Column(Integer, primary_key=True, autoincrement=True, comment="License ID")
    license_key = Column(
        String(255), nullable=False, unique=True, index=True, comment="License 密钥"
    )
    license_type = Column(
        String(50), nullable=False, comment="授权类型：trial/standard/enterprise"
    )
    max_users = Column(
        Integer, nullable=True, comment="最大用户数（NULL 表示不限制）"
    )
    authorized_modules = Column(
        JSON, nullable=True, comment="授权模块列表（JSON 数组）"
    )
    machine_code = Column(
        String(255), nullable=True, comment="绑定的机器码（NULL 表示不绑定）"
    )
    issued_at = Column(DateTime, nullable=False, comment="签发时间")
    expires_at = Column(
        DateTime, nullable=False, index=True, comment="过期时间"
    )
    is_active = Column(
        SmallInteger, nullable=False, default=1, comment="是否激活：1-是 0-否"
    )
    activated_at = Column(DateTime, nullable=True, comment="激活时间")
    created_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
        comment="创建时间",
    )
    updated_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
        onupdate=_utcnow_naive,
        comment="更新时间",
    )
    # v1.2（R-3）：部分唯一索引，保证同一时刻最多一条 is_active=1 记录。
    # SQLite（≥3.8.0）与 PostgreSQL（≥9.5）均支持部分索引。
    __table_args__ = (
        Index(
            ACTIVE_LICENSE_UNIQUE_INDEX,
            "is_active",
            unique=True,
            sqlite_where=text("is_active = 1"),
            postgresql_where=text("is_active = 1"),
        ),
    )
