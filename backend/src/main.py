import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from src.core.config import settings
from src.core.database import AsyncSessionLocal, engine, init_db
from src.core.exceptions import setup_exception_handlers
from src.core.logging import setup_logging
from src.core.middleware import RequestIdMiddleware
from src.core.response import success_response
# 导入模型，确保 Base.metadata 包含表定义
import src.core.models  # noqa: F401
from src.modules.auth import models as auth_models  # noqa: F401
from src.modules.auth.router import router as auth_router
from src.modules.auth.service import ensure_auth_seed_data
# 初始化日志
setup_logging()
logger = logging.getLogger(__name__)
@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理。"""
    await init_db()
    logger.info("数据库连接初始化完成")
    # 初始化认证种子数据（幂等，不破坏已授权权限）。
    # 若尚未执行迁移，仅记录警告，不阻塞启动。
    try:
        async with AsyncSessionLocal() as session:
            await ensure_auth_seed_data(session)
        logger.info("认证种子数据检查完成")
    except Exception as exc:
        logger.warning(
            "认证种子数据初始化失败（请确认已执行数据库迁移）: %s", exc
        )
    yield
    await engine.dispose()
    logger.info("数据库引擎已关闭")
app = FastAPI(
    title=settings.APP_NAME,
    version="0.1.0",
    description="统一权限管理平台 API",
    lifespan=lifespan,
)
# ---------------------------------------------------------------------------
# CORS：从 settings 读取允许的源，避免硬编码
# ---------------------------------------------------------------------------
_origins = settings.cors_origins_list
app.add_middleware(
    CORSMiddleware,
    allow_origins=_origins,
    # 当源为 "*" 时浏览器不允许携带凭证，因此自动关闭 credentials
    allow_credentials=("*" not in _origins),
    allow_methods=["*"],
    allow_headers=["*"],
)
# 请求 ID 中间件
app.add_middleware(RequestIdMiddleware)
# 全局异常处理
setup_exception_handlers(app)
@app.get("/health", tags=["System"])
async def health_check():
    """服务健康检查。"""
    db_ok = False
    try:
        from sqlalchemy import text
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
            db_ok = True
    except Exception as exc:
        logger.warning("数据库连接检查失败: %s", exc)
    return success_response(
        data={
            "status": "ok",
            "database": "connected" if db_ok else "unavailable",
        }
    )
# 注册认证模块路由
app.include_router(auth_router)
