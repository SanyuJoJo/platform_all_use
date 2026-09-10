"""
Pytest 全局配置（v1.3）
v1.3 变更：
- P2-2：session 开始时若 KEEP_TEST_DB != 1 则删除旧 test_app.db，
        再通过 Base.metadata.create_all 重建，确保 schema 与模型一致。
- P2-6：新增 KEEP_TEST_DB 环境变量开关，默认清理 test_app.db。
v1.2 引入：
- P2-6：测试数据库隔离（默认 test_app.db）。
关键点：
    必须在本文件顶部、任何 src.* 模块被导入之前设置 DATABASE_URL
    环境变量，因为 src.core.config.settings 在首次导入时即固化配置。
"""
import os
# ---- 必须在导入 src 之前设置 ----
_TEST_DB_URL = os.environ.get(
    "TEST_DATABASE_URL",
    "sqlite+aiosqlite:///./test_app.db",
)
os.environ["DATABASE_URL"] = _TEST_DB_URL
# 测试库清理开关（默认清理）
_KEEP_TEST_DB = os.environ.get("KEEP_TEST_DB", "") == "1"
# 从 URL 中提取文件路径（仅对 sqlite 有效）
_TEST_DB_PATH = _TEST_DB_URL.split("///")[-1] if "sqlite" in _TEST_DB_URL else ""
# ---- 现在可以安全导入 src ----
import pytest  # noqa: E402
from src.core.database import AsyncSessionLocal, Base, engine  # noqa: E402
from src.modules.auth.service import ensure_auth_seed_data  # noqa: E402
@pytest.fixture(scope="session", autouse=True)
async def setup_test_database():
    """
    会话级夹具：
    1. （P2-2）session 开始时若 KEEP_TEST_DB != 1，删除旧 test_app.db，
       确保 schema 与模型一致。
    2. 通过 Base.metadata.create_all 快速建表。
    3. 初始化种子数据（幂等）。
    4. （P2-6）session 结束时若 KEEP_TEST_DB != 1，删除 test_app.db。
    说明：
        生产环境的迁移由 Alembic 管理。测试库通过 Base.metadata.create_all
        快速建表，避免每次运行测试都走完整的迁移链。
        若需测试迁移本身（tests/test_migrations.py），请在开发库上运行，
        不要在此夹具中处理。
    """
    # P2-2：删除旧测试库，确保 schema 与模型一致
    if not _KEEP_TEST_DB and _TEST_DB_PATH and os.path.exists(_TEST_DB_PATH):
        try:
            os.remove(_TEST_DB_PATH)
        except OSError:
            # 文件被占用等极端情况，忽略，交由 create_all 处理
            pass
    # 延迟导入所有模型，确保 Base.metadata 完整
    import src.core.models  # noqa: F401
    from src.modules.auth import models as _auth_models  # noqa: F401
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    async with AsyncSessionLocal() as session:
        await ensure_auth_seed_data(session)
    yield
    # 关闭引擎
    await engine.dispose()
    # P2-6：删除测试库
    if not _KEEP_TEST_DB and _TEST_DB_PATH and os.path.exists(_TEST_DB_PATH):
        try:
            os.remove(_TEST_DB_PATH)
        except OSError:
            pass
