"""前端静态资源托管服务。
设计要点：
    1. 主应用挂载于 "/"（SPA 兜底）；
    2. 子应用挂载于 "/sub-apps/{module_id}"，路径更具体、先注册；
    3. SPA 回退策略：
        - 仅对 Accept: text/html 的请求回退到 index.html；
        - API 路径（api/ 前缀）、健康检查路径（health 前缀）不回退；
        - 静态资源路径（assets/ 前缀）不回退，避免掩盖真实 404。
    4. 目录不存在时静默跳过（便于开发期只跑后端）。
v1.3 变更（★ P1-N4' 修复）：
    采用评审推荐的方案 A，将 health 恢复为前缀匹配：
        - NON_SPA_PREFIXES = ("api/", "health")
        - 删除空的 NON_SPA_EXACT
    效果：
        - /health        → FastAPI 显式路由命中（200 JSON）
        - /healthcheck   → 404 JSON（不参与 SPA 回退）
        - /healthz       → 404 JSON（不参与 SPA 回退）
    说明：
        - /healthcheck、/healthz 属于系统探测类路径，非前端业务路由；
        - 若未来真有 /health* 前端路由，可显式在 NON_SPA_EXACT 中作为白名单，
          当前不需要。
依赖：
    - starlette.staticfiles.StaticFiles（FastAPI 内置）
"""
from __future__ import annotations

import logging
from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from starlette.exceptions import HTTPException as StarletteHTTPException

logger = logging.getLogger(__name__)
# 前缀匹配：请求路径（去掉前导 /）以这些值开头时，不回退 SPA
NON_SPA_PREFIXES: tuple[str, ...] = (
    "api/",     # API 路由
    "health",   # 健康检查相关路径（/health、/healthcheck、/healthz 等）
)
# 静态资源前缀：不回退 SPA，避免掩盖真实 404
ASSET_PREFIX = "assets/"
def _should_skip_spa_fallback(clean_path: str) -> bool:
    """判断该路径是否应跳过 SPA 回退。
    Args:
        clean_path: 去掉前导 / 后的请求路径。
    """
    if any(clean_path.startswith(p) for p in NON_SPA_PREFIXES):
        return True
    if clean_path.startswith(ASSET_PREFIX):
        return True
    return False
class SPAStaticFiles(StaticFiles):
    """静态文件服务，未命中时按规则回退到 index.html。
    回退条件（全部满足）：
        1. 触发 404；
        2. 请求路径不以 NON_SPA_PREFIXES 中任何前缀开头
           （即不属于 API 或健康检查）；
        3. 请求路径不以 assets/ 开头；
        4. 请求头包含 Accept: text/html。
    """
    async def get_response(self, path: str, scope):  # type: ignore[override]
        try:
            return await super().get_response(path, scope)
        except StarletteHTTPException as exc:
            if exc.status_code != 404:
                raise
            clean_path = path.lstrip("/")
            # ---- 规则 2 & 3：API / 健康检查 / 静态资源不回退 ----
            if _should_skip_spa_fallback(clean_path):
                raise
            # ---- 规则 4：仅对 HTML 请求回退 ----
            headers = dict(scope.get("headers") or [])
            accept = headers.get(b"accept", b"").decode("latin-1")
            if "text/html" not in accept:
                raise
            return await super().get_response("index.html", scope)
def mount_frontend(
    app: FastAPI,
    frontend_dir: Path,
    *,
    api_prefix: str = "/api/v1",
) -> None:
    """将打包后的前端静态资源挂载到 FastAPI 应用。
    Args:
        app: FastAPI 应用实例。
        frontend_dir: 前端产物根目录（包含 main-app/ 与 sub-apps/）。
        api_prefix: API 前缀，用于日志提示。
    """
    frontend_dir = Path(frontend_dir)
    if not frontend_dir.exists():
        logger.warning(
            "前端产物目录不存在，跳过静态资源挂载: %s "
            "（开发期可只运行后端，或先执行 `pnpm build:package`）",
            frontend_dir,
        )
        return
    sub_apps_dir = frontend_dir / "sub-apps"
    main_app_dir = frontend_dir / "main-app"
    mounted_sub_apps: list[str] = []
    # ---- 1. 先挂载子应用（更具体的路径，需优先注册）----
    if sub_apps_dir.exists():
        for sub_dir in sorted(sub_apps_dir.iterdir()):
            if not sub_dir.is_dir():
                continue
            index_html = sub_dir / "index.html"
            if not index_html.exists():
                logger.warning("子应用目录缺少 index.html，跳过: %s", sub_dir)
                continue
            mount_path = f"/sub-apps/{sub_dir.name}"
            app.mount(
                mount_path,
                SPAStaticFiles(directory=str(sub_dir), html=True),
                name=f"subapp-{sub_dir.name}",
            )
            mounted_sub_apps.append(sub_dir.name)
    # ---- 2. 再挂载主应用（根路径，作为 SPA 兜底）----
    if main_app_dir.exists() and (main_app_dir / "index.html").exists():
        app.mount(
            "/",
            SPAStaticFiles(directory=str(main_app_dir), html=True),
            name="main-app",
        )
        logger.info("已挂载主应用: /  →  %s", main_app_dir)
    else:
        logger.warning(
            "主应用产物缺失，未挂载根路径: %s",
            main_app_dir / "index.html",
        )
    if mounted_sub_apps:
        logger.info(
            "已挂载 %d 个子应用: %s",
            len(mounted_sub_apps),
            ", ".join(mounted_sub_apps),
        )
    logger.info(
        "前端资源托管已就绪 (api_prefix=%s, frontend_dir=%s)",
        api_prefix,
        frontend_dir,
    )
