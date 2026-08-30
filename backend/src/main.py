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
# 导入所有模型，以便 Base.metadata 包含表定义
# 注：当开发模块时，在此处导入 models
# from modules.auth import models  # 示例
import src.core.models  # 基础示例模型
# 初始化日志
setup_logging()
logger = logging.getLogger(__name__)
@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动时初始化数据库（仅验证连接，表由迁移创建）
    await init_db()
    logger.info("数据库连接初始化完成")
    yield
    # 关闭时清理资源
    await engine.dispose()
    logger.info("数据库引擎已关闭")
app = FastAPI(
    title=settings.APP_NAME,
    version="0.1.0",
    description="统一权限管理平台 API",
    lifespan=lifespan,
)
# 配置 CORS（开发允许所有源）
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
# 注册请求 ID 中间件（需在异常处理之前）
app.add_middleware(RequestIdMiddleware)
# 注册全局异常处理器
setup_exception_handlers(app)
# ---------- 健康检查 ----------
@app.get("/health", tags=["System"])
async def health_check():
    """服务健康检查，包含数据库连接状态"""
    db_ok = False
    try:
        from sqlalchemy import text
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
            db_ok = True
    except Exception as e:
        logger.warning(f"数据库连接检查失败: {e}")
    return success_response(data={"status": "ok", "database": "connected" if db_ok else "unavailable"})
# ---------- 后续可注册模块路由 ----------
# from modules.auth.router import router as auth_router
# app.include_router(auth_router, prefix="/api/v1/auth", tags=["Auth"])
