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
import src.core.models
from src.modules.auth import models as auth_models
from src.modules.module_manager import models as module_manager_models
from src.modules.audit_log import models as audit_log_models
from src.modules.license import models as license_models  # noqa: F401
# ==========================================
config = context.config
if config.config_file_name is not None:
    fileConfig(config.config_file_name)
from src.core.config import settings
config.set_main_option("sqlalchemy.url", settings.DATABASE_URL)
target_metadata = Base.metadata
def run_migrations_offline() -> None:
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
