"""数据库引擎与会话管理。"""
from sqlalchemy import event, text
from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import declarative_base
from src.core.config import settings
engine = create_async_engine(
    settings.DATABASE_URL,
    echo=settings.DEBUG,
    future=True,
)
# ---------------------------------------------------------------------------
# SQLite 外键启用（v1.1 P1-2）
# ---------------------------------------------------------------------------
@event.listens_for(engine.sync_engine, "connect")
def _set_sqlite_pragma(dbapi_connection, connection_record):
    """
    每次建立 SQLite 连接时执行 PRAGMA foreign_keys=ON。
    原因：SQLite 默认不启用外键约束，会导致 auth_user_role /
    auth_refresh_token 等表的 ON DELETE CASCADE 失效，删除用户时残留
    关联记录。
    对 PostgreSQL 等数据库不生效。
    """
    if settings.DATABASE_URL.startswith("sqlite"):
        cursor = dbapi_connection.cursor()
        cursor.execute("PRAGMA foreign_keys=ON")
        cursor.close()
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
    autocommit=False,
    autoflush=False,
)
Base = declarative_base()
async def init_db():
    """初始化数据库（仅测试连接，表结构由迁移管理）。"""
    async with engine.connect() as conn:
        await conn.execute(text("SELECT 1"))
async def get_db() -> AsyncSession:
    """依赖注入：获取数据库会话。"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()
