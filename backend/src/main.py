import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from src.core.config import settings
from src.core.database import init_db, engine
from src.core.exceptions import setup_exception_handlers
from src.core.response import success_response
from src.core.logging import setup_logging
from src.core.middleware import RequestIdMiddleware
import src.core.models
setup_logging()
logger = logging.getLogger(__name__)
@asynccontextmanager
async def lifespan(app: FastAPI):
    await init_db()
    logger.info("数据库连接初始化完成")
    yield
    await engine.dispose()
    logger.info("数据库引擎已关闭")
app = FastAPI(
    title=settings.APP_NAME,
    version="0.1.0",
    description="统一权限管理平台 API",
    lifespan=lifespan,
)
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins_list,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
app.add_middleware(RequestIdMiddleware)
setup_exception_handlers(app)
@app.get("/health", tags=["System"])
async def health_check():
    db_ok = False
    try:
        from sqlalchemy import text
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
            db_ok = True
    except Exception as e:
        logger.warning(f"数据库连接检查失败: {e}")
    return success_response(data={"status": "ok", "database": "connected" if db_ok else "unavailable"})
# ---------- 后续注册模块路由 ----------
# from modules.auth.router import router as auth_router
# app.include_router(auth_router, prefix="/api/v1/auth", tags=["Auth"])
