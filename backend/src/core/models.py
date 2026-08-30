from sqlalchemy import Column, Integer, String, DateTime
from sqlalchemy.sql import func
from src.core.database import Base
class SystemConfig(Base):
    """
    系统配置表（示例模型，用于演示迁移机制）
    开发者可根据实际需求扩展或替换
    """
    __tablename__ = "sys_config"
    id = Column(Integer, primary_key=True, index=True)
    key = Column(String(64), unique=True, index=True, nullable=False, comment="配置键")
    value = Column(String(512), nullable=False, comment="配置值")
    description = Column(String(255), comment="描述")
    created_at = Column(DateTime, server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime, onupdate=func.now(), comment="更新时间")
