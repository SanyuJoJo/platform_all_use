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
from src.modules.module_manager import models as module_manager_models  # noqa: F401
from src.modules.audit_log import models as audit_log_models  # noqa: F401
# 认证模块路由
from src.modules.auth.router import router as auth_router
from src.modules.auth.service import ensure_auth_seed_data
from src.modules.auth.user_router import router as user_router
from src.modules.auth.role_router import router as role_router
from src.modules.auth.permission_router import router as permission_router
# 模块管理模块路由
from src.modules.module_manager.router import router as module_manager_router
from src.modules.module_manager.service import ensure_module_seed_data
from src.modules.module_manager.loader import (
    cleanup_module_residue,
    load_active_modules,
)
# 日志审计模块路由与中间件
from src.modules.audit_log.router import router as audit_log_router
from src.modules.audit_log.middleware import (
    AuditLogMiddleware,
    wait_pending_audit_tasks,
)
setup_logging()
logger = logging.getLogger(__name__)
@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理。"""
    await init_db()
    logger.info("数据库连接初始化完成")
    # V12-P1-07：启动时清理模块目录残留
    try:
        cleanup_module_residue()
    except Exception as exc:
        logger.warning("清理模块残留失败：%s", exc)
    try:
        async with AsyncSessionLocal() as session:
            await ensure_auth_seed_data(session)
        logger.info("认证种子数据检查完成")
    except Exception as exc:
        logger.warning("认证种子数据初始化失败：%s", exc)
    try:
        async with AsyncSessionLocal() as session:
            await ensure_module_seed_data(session)
            await load_active_modules(app, session)
        app.openapi_schema = None
        logger.info("模块种子数据与动态加载完成")
    except Exception as exc:
        logger.warning("模块初始化失败：%s", exc)
    yield
    # P1-2：关闭前等待审计日志后台任务（最多 5 秒）
    try:
        await wait_pending_audit_tasks(timeout=5.0)
    except Exception as exc:
        logger.warning("等待审计日志后台任务失败：%s", exc)
    await engine.dispose()
    logger.info("数据库引擎已关闭")
app = FastAPI(
    title=settings.APP_NAME,
    version="0.1.0",
    description="统一权限管理平台 API",
    lifespan=lifespan,
)
# ---------------------------------------------------------------------------
# 中间件注册顺序（Starlette LIFO：add_middleware 是 insert(0)）：
#   请求 → RequestIdMiddleware → AuditLogMiddleware → CORSMiddleware → 路由
#   响应 ← RequestIdMiddleware ← AuditLogMiddleware ← CORSMiddleware ← 路由
#
# 注意：最后添加的最先执行（请求路径）。
# 先添加 AuditLogMiddleware，再添加 RequestIdMiddleware，
# 保证 RequestIdMiddleware 在请求路径上先执行，request_id 已就绪。
# ---------------------------------------------------------------------------
_origins = settings.cors_origins_list
app.add_middleware(
    CORSMiddleware,
    allow_origins=_origins,
    allow_credentials=("*" not in _origins),
    allow_methods=["*"],
    allow_headers=["*"],
)
app.add_middleware(AuditLogMiddleware)
app.add_middleware(RequestIdMiddleware)
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
# ---------------------------------------------------------------------------
# 注册核心模块路由
# ---------------------------------------------------------------------------
app.include_router(auth_router)              # /api/v1/auth/*
app.include_router(user_router)              # /api/v1/auth/users/*
app.include_router(role_router)              # /api/v1/auth/roles/*
app.include_router(permission_router)        # /api/v1/auth/permissions/*
app.include_router(module_manager_router)    # /api/v1/modules/*
app.include_router(audit_log_router)         # /api/v1/audit-logs/*
