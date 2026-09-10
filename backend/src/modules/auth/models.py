"""
认证授权模块 - 数据库模型
表结构依据《数据库设计文档》v1.0 第二章、第三章。
"""
from sqlalchemy import (
    Column,
    DateTime,
    ForeignKey,
    Integer,
    SmallInteger,
    String,
    UniqueConstraint,
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from src.core.database import Base
class User(Base):
    __tablename__ = "auth_user"
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), nullable=False, unique=True, index=True)
    password_hash = Column(String(255), nullable=False)
    nickname = Column(String(50), nullable=False)
    email = Column(String(100), nullable=True, index=True)
    avatar = Column(String(255), nullable=True)
    status = Column(SmallInteger, nullable=False, default=1, comment="1-启用 0-禁用")
    last_login_at = Column(DateTime, nullable=True)
    last_login_ip = Column(String(45), nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, onupdate=func.now())
    # 关系（不设置 ORM 级联，数据库外键已处理中间表删除）
    roles = relationship("Role", secondary="auth_user_role", back_populates="users")
class Role(Base):
    __tablename__ = "auth_role"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), nullable=False)
    code = Column(String(50), nullable=False, unique=True, index=True)
    description = Column(String(255), nullable=True)
    is_system = Column(SmallInteger, nullable=False, default=0, comment="1-系统内置")
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, onupdate=func.now())
    users = relationship("User", secondary="auth_user_role", back_populates="roles")
    permissions = relationship(
        "Permission", secondary="auth_role_permission", back_populates="roles"
    )
class Permission(Base):
    __tablename__ = "auth_permission"
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(100), nullable=False, unique=True, index=True)
    name = Column(String(50), nullable=False)
    module_id = Column(String(50), nullable=False, index=True)
    resource = Column(String(50), nullable=False)
    action = Column(String(50), nullable=False)
    created_at = Column(DateTime, server_default=func.now())
    roles = relationship(
        "Role", secondary="auth_role_permission", back_populates="permissions"
    )
class UserRole(Base):
    __tablename__ = "auth_user_role"
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(
        Integer, ForeignKey("auth_user.id", ondelete="CASCADE"), nullable=False
    )
    role_id = Column(
        Integer, ForeignKey("auth_role.id", ondelete="CASCADE"), nullable=False
    )
    created_at = Column(DateTime, server_default=func.now())
    __table_args__ = (
        UniqueConstraint("user_id", "role_id", name="uq_auth_user_role_user_role"),
    )
class RolePermission(Base):
    __tablename__ = "auth_role_permission"
    id = Column(Integer, primary_key=True, index=True)
    role_id = Column(
        Integer, ForeignKey("auth_role.id", ondelete="CASCADE"), nullable=False
    )
    permission_id = Column(
        Integer, ForeignKey("auth_permission.id", ondelete="CASCADE"), nullable=False
    )
    created_at = Column(DateTime, server_default=func.now())
    __table_args__ = (
        UniqueConstraint(
            "role_id", "permission_id", name="uq_auth_role_permission_role_perm"
        ),
    )
class RefreshToken(Base):
    """
    Refresh Token 持久化表，支持撤销与失效校验。
    - token_hash：SHA-256 摘要，避免明文落库；
    - revoked：1 表示已撤销（登出、轮换、改密、强制下线）；
    - updated_at：符合数据库设计文档通用字段要求。
    """
    __tablename__ = "auth_refresh_token"
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(
        Integer,
        ForeignKey("auth_user.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    token_hash = Column(String(64), nullable=False, unique=True, index=True)
    expires_at = Column(DateTime, nullable=False, index=True)
    revoked = Column(SmallInteger, nullable=False, default=0, comment="1-已撤销 0-有效")
    revoked_at = Column(DateTime, nullable=True)
    ip = Column(String(45), nullable=True)
    user_agent = Column(String(255), nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, onupdate=func.now())
