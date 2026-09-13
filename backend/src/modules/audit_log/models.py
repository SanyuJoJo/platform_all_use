"""日志审计模块 - ORM 模型。"""
from datetime import datetime, timezone
from sqlalchemy import (
    Column,
    DateTime,
    ForeignKey,
    Integer,
    String,
    Text,
)
from sqlalchemy.sql import func
from src.core.database import Base
def _utcnow_naive() -> datetime:
    """UTC 当前时间（naive），与 SQLite DateTime 存储格式一致。"""
    return datetime.now(timezone.utc).replace(tzinfo=None)
class AuditLogOperation(Base):
    """
    操作日志表（audit_log_operation）。
    严格遵循《数据库设计文档》§ 5.1。
    审计日志不可修改，因此不含 updated_at 字段。
    """
    __tablename__ = "audit_log_operation"
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(
        Integer,
        ForeignKey("auth_user.id", ondelete="SET NULL"),
        nullable=True,
        index=True,
        comment="操作用户 ID",
    )
    username = Column(String(50), nullable=True, comment="操作用户名（冗余）")
    module_id = Column(
        String(50), nullable=False, index=True, comment="所属模块 ID"
    )
    action = Column(
        String(50), nullable=False, index=True, comment="操作类型"
    )
    resource = Column(String(50), nullable=True, comment="操作资源类型")
    resource_id = Column(String(50), nullable=True, comment="操作资源 ID")
    detail = Column(Text, nullable=True, comment="操作详情")
    ip = Column(String(45), nullable=True, comment="客户端 IP")
    user_agent = Column(String(255), nullable=True, comment="客户端 User-Agent")
    status = Column(
        String(20), nullable=False, default="success", comment="success/fail"
    )
    error_code = Column(Integer, nullable=True, comment="错误码（失败时）")
    request_id = Column(String(36), nullable=True, comment="请求追踪 ID")
    created_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
        index=True,
        comment="创建时间",
    )
