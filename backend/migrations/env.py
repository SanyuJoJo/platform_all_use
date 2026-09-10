import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))
import asyncio
from logging.config import fileConfig
from sqlalchemy import pool
from sqlalchemy.engine import Connection
from sqlalchemy.ext.asyncio import async_engine_from_config
from alembic import context
# ========== 导入 Base 和所有模型 ==========
from src.core.database import Base
# 导入示例模型（原有）
import src.core.models
# 导入认证模块模型
from src.modules.auth import models as auth_models
# 后续新增模块时，在此处导入即可
# from src.modules.module_manager import models as module_models
# from src.modules.audit_log import models as audit_models
# ==========================================
# Alembic Config 对象
config = context.config
if config.config_file_name is not None:
    fileConfig(config.config_file_name)
# 从环境变量获取数据库 URL
from src.core.config import settings
config.set_main_option("sqlalchemy.url", settings.DATABASE_URL)
# 目标元数据（包含所有导入的模型）
target_metadata = Base.metadata
def run_migrations_offline() -> None:
    """离线模式运行迁移"""
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )
    with context.begin_transaction():
        context.run_migrations()
def do_run_migrations(connection: Connection) -> None:
    context.configure(connection=connection, target_metadata=target_metadata)
    with context.begin_transaction():
        context.run_migrations()
async def run_async_migrations() -> None:
    """在线模式（异步）运行迁移"""
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )
    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)
    await connectable.dispose()
def run_migrations_online() -> None:
    asyncio.run(run_async_migrations())
if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
