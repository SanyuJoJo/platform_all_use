"""
验证迁移环境是否正常
"""
import pytest
from sqlalchemy import inspect, text
from src.core.database import engine
@pytest.mark.asyncio
async def test_tables_created():
    """检查核心表是否已创建（使用 run_sync 包装 inspect）"""
    async with engine.connect() as conn:
        def _get_tables(sync_conn):
            inspector = inspect(sync_conn)
            return inspector.get_table_names()
        tables = await conn.run_sync(_get_tables)
        expected_tables = [
            'auth_user', 'auth_role', 'auth_permission',
            'auth_user_role', 'auth_role_permission'
        ]
        for table in expected_tables:
            assert table in tables, f"表 {table} 未创建"
@pytest.mark.asyncio
async def test_migration_version():
    """检查迁移版本表是否存在"""
    async with engine.connect() as conn:
        result = await conn.execute(text("SELECT name FROM sqlite_master WHERE type='table' AND name='alembic_version'"))
        row = result.fetchone()
        assert row is not None, "alembic_version 表不存在"
        version = await conn.execute(text("SELECT version_num FROM alembic_version"))
        assert version.fetchone() is not None, "迁移版本号未记录"
