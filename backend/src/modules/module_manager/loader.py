"""模块动态加载器（v1.4）。"""
import importlib
import logging
from pathlib import Path
from typing import Set

from fastapi import Depends, FastAPI
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from src.core.config import settings
from src.core.database import get_db
from src.core.exceptions import PlatformException
from src.modules.module_manager.constants import CORE_MODULE_IDS
from src.modules.module_manager.models import Module
from src.modules.module_manager.manifest import load_manifest_from_dir

logger = logging.getLogger(__name__)


def get_modules_root() -> Path:
    """获取模块源码根目录。"""
    root = Path(settings.MODULES_DIR)
    if not root.is_absolute():
        root = Path.cwd() / root
    return root.resolve()


def get_module_dir(module_id: str) -> Path:
    """获取业务模块目录：src/modules/module_{module_id}"""
    return get_modules_root() / f"module_{module_id}"


# ---------------------------------------------------------------------------
# 已加载集合统一管理
# ---------------------------------------------------------------------------
def _get_loaded_modules(app: FastAPI) -> Set[str]:
    """获取已加载模块集合（惰性初始化；核心模块默认已加载）。"""
    if not hasattr(app.state, "_loaded_modules"):
        app.state._loaded_modules = set(CORE_MODULE_IDS)
    return app.state._loaded_modules


def _mark_loaded(app: FastAPI, module_id: str) -> None:
    _get_loaded_modules(app).add(module_id)


def _unmark_loaded(app: FastAPI, module_id: str) -> None:
    _get_loaded_modules(app).discard(module_id)


def _is_loaded(app: FastAPI, module_id: str) -> bool:
    return module_id in _get_loaded_modules(app)


def unmark_module_loaded(app: FastAPI, module_id: str) -> None:
    """对外暴露：从已加载集合中移除模块。"""
    _unmark_loaded(app, module_id)


# ---------------------------------------------------------------------------
# 启动残留清理（V13-P1-04：增加 .bak）
# ---------------------------------------------------------------------------
_RESIDUE_SUFFIXES = (".new", ".old", ".bak")


def cleanup_module_residue() -> None:
    """
    扫描模块根目录，清理安装/升级中断残留的 *.new / *.old / *.bak 目录。
    在 lifespan 启动时调用。
    """
    root = get_modules_root()
    if not root.exists():
        return
    cleaned: list[str] = []
    for child in root.iterdir():
        if not child.is_dir():
            continue
        name = child.name
        if not name.startswith("module_"):
            continue
        if any(name.endswith(suffix) for suffix in _RESIDUE_SUFFIXES):
            try:
                import shutil

                shutil.rmtree(child)
                cleaned.append(name)
            except OSError as exc:
                logger.warning("清理残留目录失败：%s，原因：%s", child, exc)
    if cleaned:
        logger.info("已清理模块残留目录：%s", cleaned)


# ---------------------------------------------------------------------------
# 路由导入与注册
# ---------------------------------------------------------------------------
def _import_router(module_id: str, entry_backend: str):
    """导入模块路由（支持 `router` 与 `api.router` 点号路径）。"""
    if ":" not in entry_backend:
        raise PlatformException(
            code=30008, message="模块入口文件配置无效", status_code=400
        )
    module_attr, router_attr = entry_backend.split(":", 1)
    package_path = f"src.modules.module_{module_id}.{module_attr}"
    try:
        mod = importlib.import_module(package_path)
    except ImportError as exc:
        raise PlatformException(
            code=30008,
            message=f"模块入口文件不存在或导入失败：{package_path}",
            status_code=400,
        ) from exc
    router = getattr(mod, router_attr, None)
    if router is None:
        raise PlatformException(
            code=30009,
            message=f"模块路由对象不存在：{package_path}.{router_attr}",
            status_code=500,
        )
    return router


def require_module_enabled(module_id: str):
    """依赖工厂：校验模块是否存在且已启用。"""

    async def _check(db: AsyncSession = Depends(get_db)):
        module = await db.get(Module, module_id)
        if not module or module.status != "active":
            raise PlatformException(
                code=90002,
                message="模块不存在或未启用",
                status_code=404,
            )

    return _check


def _register_module_router(app: FastAPI, module_id: str, router) -> None:
    """注册模块路由到 FastAPI。"""
    app.include_router(
        router,
        prefix=f"/api/v1/{module_id}",
        tags=[module_id],
        dependencies=[Depends(require_module_enabled(module_id))],
    )
    app.openapi_schema = None


# ---------------------------------------------------------------------------
# 启动加载
# ---------------------------------------------------------------------------
async def load_active_modules(app: FastAPI, db: AsyncSession) -> None:
    """
    启动时加载所有 active 业务模块。

    v1.4 变更：
    - V13-P1-07：manifest_updated 重命名为 needs_commit。
    """
    result = await db.execute(
        select(Module)
        .options(selectinload(Module.dependencies))
        .where(Module.status == "active")
    )
    modules = list(result.scalars().all())

    modules_map = {m.id: m for m in modules}
    sorted_modules: list[Module] = []
    visited: set[str] = set()

    def visit(m: Module) -> None:
        if m.id in visited:
            return
        visited.add(m.id)
        for dep in m.dependencies:
            dep_module = modules_map.get(dep.dependency_id)
            if dep_module is not None:
                visit(dep_module)
        sorted_modules.append(m)

    for m in modules:
        visit(m)

    loaded = _get_loaded_modules(app)
    needs_commit = False

    for module in sorted_modules:
        if module.id in CORE_MODULE_IDS:
            continue

        missing_deps = [
            d.dependency_id
            for d in module.dependencies
            if d.dependency_id not in loaded
        ]
        if missing_deps:
            logger.error(
                "跳过模块 %s：依赖未加载 %s", module.id, missing_deps
            )
            continue

        try:
            if not module.manifest:
                module_dir = get_module_dir(module.id)
                module.manifest = load_manifest_from_dir(module_dir)
                needs_commit = True

            router = _import_router(module.id, module.entry_backend)
            _register_module_router(app, module.id, router)
            _mark_loaded(app, module.id)
            logger.info("模块 %s 路由加载成功", module.id)
        except Exception as exc:
            logger.error("模块 %s 加载失败：%s", module.id, exc)

    if needs_commit:
        await db.commit()


# ---------------------------------------------------------------------------
# 热加载
# ---------------------------------------------------------------------------
async def load_single_module(
    app: FastAPI, db: AsyncSession, module_id: str
) -> None:
    """
    运行时热加载单个模块路由（幂等）。

    若模块已在 _loaded_modules 中，直接返回。
    若路由导入失败，抛出 PlatformException。
    """
    if _is_loaded(app, module_id):
        logger.debug("模块 %s 已加载，跳过", module_id)
        return

    module = await db.get(Module, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    if module.id in CORE_MODULE_IDS:
        _mark_loaded(app, module_id)
        return

    router = _import_router(module.id, module.entry_backend)
    _register_module_router(app, module.id, router)
    _mark_loaded(app, module_id)
    logger.info("模块 %s 热加载成功", module.id)
