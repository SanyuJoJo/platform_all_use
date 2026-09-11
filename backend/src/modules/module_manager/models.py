"""模块管理模块 - ORM 模型。"""
from datetime import datetime, timezone

from sqlalchemy import (
    Column,
    DateTime,
    ForeignKey,
    Integer,
    JSON,
    String,
    Text,
    UniqueConstraint,
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from src.core.database import Base


def _utcnow_naive() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


class Module(Base):
    """模块主表。"""
    __tablename__ = "module_manager_module"

    id = Column(String(50), primary_key=True, comment="模块 ID")
    name = Column(String(50), nullable=False, comment="模块名称")
    version = Column(String(20), nullable=False, comment="语义化版本")
    description = Column(Text, nullable=True, comment="模块描述")
    author = Column(String(100), nullable=True, comment="作者")
    homepage = Column(String(255), nullable=True, comment="主页")
    status = Column(
        String(20), nullable=False, default="inactive", comment="active/inactive"
    )
    entry_backend = Column(String(100), nullable=False, comment="后端入口，如 router:router")
    entry_frontend = Column(String(255), nullable=True, comment="前端入口 URL")
    config = Column(JSON, nullable=True, default=dict, comment="模块配置")
    manifest = Column(JSON, nullable=True, comment="完整模块清单（含 menus/permissions/config_schema）")
    installed_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
    )
    updated_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
        onupdate=_utcnow_naive,
    )

    dependencies = relationship(
        "ModuleDependency",
        back_populates="module",
        cascade="all, delete-orphan",
        foreign_keys="ModuleDependency.module_id",
    )


class ModuleDependency(Base):
    """模块依赖表。"""
    __tablename__ = "module_manager_dependency"

    id = Column(Integer, primary_key=True, autoincrement=True)
    module_id = Column(
        String(50),
        ForeignKey("module_manager_module.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    dependency_id = Column(String(50), nullable=False, index=True)
    created_at = Column(
        DateTime,
        nullable=False,
        default=_utcnow_naive,
        server_default=func.now(),
    )

    module = relationship(
        "Module", back_populates="dependencies", foreign_keys=[module_id]
    )

    __table_args__ = (
        UniqueConstraint(
            "module_id",
            "dependency_id",
            name="uq_module_manager_dependency_module_dep",
        ),
    )
