"""应用入口。
v1.1（集成部署包）变更：
- P0-2：前端 deploy 路径改为环境变量 `FRONTEND_DEPLOY_DIR` 优先，
        默认值回退到 `parents[2] / "frontend" / "deploy"`（本地仓库正确）。
- P0-3：SPA fallback 的注册从模块导入时延后到 `lifespan` 中，
        且在 `load_active_modules` 之后执行，避免拦截动态模块路由。
- P2-4：`_SPA_EXCLUDE_PREFIXES` 改为 `"api"`，同时匹配 `/api` 与 `/api/...`。
- P2-5：静态资源未挂载时，`/assets/*`、`/sub-apps/*` 显式返回 404。
"""
import logging
import os
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

# 导入模型，确保 Base.metadata 包含表定义
import src.core.models  # noqa: F401
from src.core.config import settings
from src.core.database import AsyncSessionLocal, engine, init_db
from src.core.exceptions import setup_exception_handlers
from src.core.logging import setup_logging
from src.core.middleware import RequestIdMiddleware
from src.core.response import success_response
from src.modules.audit_log import models as audit_log_models  # noqa: F401
from src.modules.audit_log.middleware import (
    AuditLogMiddleware,
    wait_pending_audit_tasks,
)

# 日志审计模块路由与中间件
from src.modules.audit_log.router import router as audit_log_router
from src.modules.auth import models as auth_models  # noqa: F401
from src.modules.auth.permission_router import router as permission_router
from src.modules.auth.role_router import router as role_router

# 认证模块路由
from src.modules.auth.router import router as auth_router
from src.modules.auth.service import ensure_auth_seed_data
from src.modules.auth.user_router import router as user_router
from src.modules.license import models as license_models  # noqa: F401

# License 管理模块路由与启动校验
from src.modules.license.router import router as license_router
from src.modules.license.service import verify_license_on_startup
from src.modules.module_manager import models as module_manager_models  # noqa: F401
from src.modules.module_manager.loader import (
    cleanup_module_residue,
    load_active_modules,
)

# 模块管理模块路由
from src.modules.module_manager.router import router as module_manager_router
from src.modules.module_manager.service import ensure_module_seed_data

setup_logging()
logger = logging.getLogger(__name__)
# ---------------------------------------------------------------------------
# 前端静态资源路径解析（P0-2 修复）
# ---------------------------------------------------------------------------
def _resolve_frontend_deploy_dir() -> Path:
    """解析前端 deploy 目录。
    优先级：
        1. 环境变量 `FRONTEND_DEPLOY_DIR`（Docker / compose / 部署包必须显式设置）
        2. 默认值：`<repo>/frontend/deploy`
           - `__file__` = `<repo>/backend/src/main.py`
           - `parents[0]` = `<repo>/backend/src`
           - `parents[1]` = `<repo>/backend`
           - `parents[2]` = `<repo>`
    """
    env_val = os.getenv("FRONTEND_DEPLOY_DIR", "").strip()
    if env_val:
        return Path(env_val).resolve()
    return (
        Path(__file__).resolve().parents[2] / "frontend" / "deploy"
    ).resolve()
_FRONTEND_DEPLOY_DIR: Path = _resolve_frontend_deploy_dir()
_MAIN_APP_DIR: Path = _FRONTEND_DEPLOY_DIR / "main-app"
_SUB_APPS_DIR: Path = _FRONTEND_DEPLOY_DIR / "sub-apps"
_MAIN_APP_INDEX: Path = _MAIN_APP_DIR / "index.html"
_MAIN_APP_ASSETS: Path = _MAIN_APP_DIR / "assets"
# 记录静态资源挂载状态，供 SPA fallback 判断（P2-5）
_STATIC_MOUNTED: dict[str, bool] = {"assets": False, "sub_apps": False}
# ---------------------------------------------------------------------------
# 静态资源挂载（模块级执行：app.mount 使用前缀匹配，不参与路由顺序）
# ---------------------------------------------------------------------------
def _mount_static_assets(app: FastAPI) -> None:
    """挂载主应用 assets 与子应用静态资源。"""
    if not _MAIN_APP_INDEX.exists():
        logger.info(
            "未检测到前端产物（%s 不存在），跳过静态资源挂载",
            _MAIN_APP_INDEX,
        )
        return
    # 1. 主应用 assets
    if _MAIN_APP_ASSETS.exists():
        app.mount(
            "/assets",
            StaticFiles(directory=str(_MAIN_APP_ASSETS)),
            name="main-assets",
        )
        _STATIC_MOUNTED["assets"] = True
        logger.info("已挂载主应用资源目录：%s → /assets", _MAIN_APP_ASSETS)
    else:
        logger.warning("主应用 assets 目录不存在：%s", _MAIN_APP_ASSETS)
    # 2. 子应用静态资源（html=True 自动处理 index.html）
    if _SUB_APPS_DIR.exists():
        app.mount(
            "/sub-apps",
            StaticFiles(directory=str(_SUB_APPS_DIR), html=True),
            name="sub-apps",
        )
        _STATIC_MOUNTED["sub_apps"] = True
        logger.info("已挂载子应用目录：%s → /sub-apps", _SUB_APPS_DIR)
    else:
        logger.warning("子应用目录不存在：%s", _SUB_APPS_DIR)
# ---------------------------------------------------------------------------
# SPA fallback 注册（P0-3 / P2-4 / P2-5 修复）
# ---------------------------------------------------------------------------
_SPA_EXCLUDE_PREFIXES: tuple[str, ...] = (
    "api",
    "health",
    "docs",
    "redoc",
    "openapi.json",
)
def _register_spa_fallback(app: FastAPI) -> None:
    """注册 SPA fallback。
    必须在 `load_active_modules` 之后调用，否则会拦截动态模块路由（P0-3）。
    幂等：若已注册则直接返回。
    """
    if getattr(app.state, "_spa_fallback_registered", False):
        return
    if not _MAIN_APP_INDEX.exists():
        logger.info("未检测到前端产物，跳过 SPA fallback 注册")
        return
    @app.get("/{full_path:path}", include_in_schema=False)
    async def spa_fallback(full_path: str):
        # 1. 排除 API 与系统路径（同时匹配 `/api` 与 `/api/...`）
        for prefix in _SPA_EXCLUDE_PREFIXES:
            if full_path == prefix or full_path.startswith(prefix + "/"):
                raise HTTPException(status_code=404, detail="Not Found")
        # 2. 静态资源未挂载时返回 404，而非 HTML（P2-5）
        if not _STATIC_MOUNTED["assets"]:
            if full_path == "assets" or full_path.startswith("assets/"):
                raise HTTPException(status_code=404, detail="Not Found")
        if not _STATIC_MOUNTED["sub_apps"]:
            if full_path == "sub-apps" or full_path.startswith("sub-apps/"):
                raise HTTPException(status_code=404, detail="Not Found")
        # 3. 若命中主应用目录下的真实文件（如 favicon.ico），直接返回
        if full_path:
            try:
                candidate = (_MAIN_APP_DIR / full_path).resolve()
                candidate.relative_to(_MAIN_APP_DIR.resolve())  # 防路径穿越
                if candidate.is_file():
                    return FileResponse(str(candidate))
            except (ValueError, OSError):
                pass
        # 4. 其余路径统一返回 SPA index.html
        return FileResponse(str(_MAIN_APP_INDEX))
    app.state._spa_fallback_registered = True
    logger.info("已注册 SPA fallback：非 API 路径 → %s", _MAIN_APP_INDEX)
# ---------------------------------------------------------------------------
# 生命周期
# ---------------------------------------------------------------------------
@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理。
    启动顺序（关键）：
        init_db
          → cleanup_module_residue
          → ensure_auth_seed_data
          → ensure_module_seed_data
          → load_active_modules        ← 动态模块路由在此追加
          → verify_license_on_startup
          → _register_spa_fallback     ← 必须在动态模块路由之后（P0-3）
    """
    await init_db()
    logger.info("数据库连接初始化完成")
    # 启动时清理模块目录残留
    try:
        cleanup_module_residue()
    except Exception as exc:
        logger.warning("清理模块残留失败：%s", exc)
    # 认证种子数据
    try:
        async with AsyncSessionLocal() as session:
            await ensure_auth_seed_data(session)
        logger.info("认证种子数据检查完成")
    except Exception as exc:
        logger.warning("认证种子数据初始化失败：%s", exc)
    # 模块种子数据与动态加载
    try:
        async with AsyncSessionLocal() as session:
            await ensure_module_seed_data(session)
            await load_active_modules(app, session)
        app.openapi_schema = None
        logger.info("模块种子数据与动态加载完成")
    except Exception as exc:
        logger.warning("模块初始化失败：%s", exc)
    # License 启动校验（不阻断启动）
    try:
        async with AsyncSessionLocal() as session:
            result = await verify_license_on_startup(session)
            if not result.get("ok"):
                logger.warning(
                    "License 启动校验未通过：code=%s message=%s",
                    result.get("code"),
                    result.get("message"),
                )
    except Exception as exc:
        logger.warning("License 启动校验异常：%s", exc)
    # ★ P0-3 修复：SPA fallback 必须在所有动态模块路由注册之后执行
    _register_spa_fallback(app)
    yield
    # 关闭前等待审计日志后台任务（最多 5 秒）
    try:
        await wait_pending_audit_tasks(timeout=5.0)
    except Exception as exc:
        logger.warning("等待审计日志后台任务失败：%s", exc)
    await engine.dispose()
    logger.info("数据库引擎已关闭")
# ---------------------------------------------------------------------------
# 应用初始化
# ---------------------------------------------------------------------------
app = FastAPI(
    title=settings.APP_NAME,
    version="0.1.0",
    description="统一权限管理平台 API",
    lifespan=lifespan,
)
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
# 注册核心模块路由（顺序无关）
# ---------------------------------------------------------------------------
app.include_router(auth_router)              # /api/v1/auth/*
app.include_router(user_router)              # /api/v1/auth/users/*
app.include_router(role_router)              # /api/v1/auth/roles/*
app.include_router(permission_router)        # /api/v1/auth/permissions/*
app.include_router(module_manager_router)    # /api/v1/modules/*
app.include_router(audit_log_router)         # /api/v1/audit-logs/*
app.include_router(license_router)           # /api/v1/license/*
# ---------------------------------------------------------------------------
# 静态资源挂载（模块级执行）
#
# 说明：
#   - `/assets` 与 `/sub-apps` 使用 app.mount（前缀匹配），不参与路由顺序，
#     可以安全地在模块级执行；
#   - SPA fallback 是 catch-all Route，必须在 lifespan 中延后注册（P0-3）；
#   - 未检测到前端产物时，全部跳过，不影响后端独立运行。
# ---------------------------------------------------------------------------
_mount_static_assets(app)
