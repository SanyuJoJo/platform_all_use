"""日志审计模块 - 全局中间件。
自动记录所有进入 /api/v1/* 的请求到 audit_log_operation 表。
v1.1 变更：
- P1-1：单段路径 resource 解析为 path_module，不再返回 None。
- P1-2：改为 asyncio.create_task 异步调度写入任务，不阻塞响应。
        提供 wait_pending_audit_tasks 供 lifespan 关闭时等待。
- P1-3：优先从响应头 X-Error-Code 读取精确错误码，
        回退到 STATUS_ERROR_MAP 粗略映射。
- P1-5：路径以 /export 结尾时 action="export"。
"""
import asyncio
import logging
from typing import Optional, Set, Tuple

from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import Response

from src.core.logging import request_id_var
from src.core.security import decode_token
from src.modules.audit_log.constants import (
    METHOD_ACTION_MAP,
    PATH_MODULE_MAP,
    SKIP_PATH_PREFIXES,
    STATUS_ERROR_MAP,
)
from src.modules.audit_log.service import write_operation_log

logger = logging.getLogger(__name__)
# 模块级后台任务集合：保存 asyncio.Task 引用，防止被 GC 回收
_pending_audit_tasks: Set[asyncio.Task] = set()
# ---------------------------------------------------------------------------
# 后台任务管理
# ---------------------------------------------------------------------------
def _schedule_audit_write(coro) -> None:
    """
    将审计日志写入协程调度为后台任务。
    - 保存 Task 引用防止被 GC；
    - 任务完成时自动从集合中移除；
    - 若事件循环未运行（极端情况），降级为记录 ERROR 日志。
    """
    try:
        task = asyncio.create_task(coro)
    except RuntimeError as exc:
        # 事件循环未运行（如单元测试同步上下文）
        logger.error("调度审计日志后台任务失败（事件循环未运行）：%s", exc)
        return
    _pending_audit_tasks.add(task)
    task.add_done_callback(_pending_audit_tasks.discard)
async def wait_pending_audit_tasks(timeout: float = 5.0) -> None:
    """
    等待所有未完成的审计日志写入任务（供 lifespan 关闭时调用）。
    参数：
        timeout: 最长等待时间（秒），默认 5 秒。超时后仅记录 WARNING，
                 不阻塞进程退出。
    """
    if not _pending_audit_tasks:
        return
    tasks = list(_pending_audit_tasks)
    try:
        await asyncio.wait_for(
            asyncio.gather(*tasks, return_exceptions=True),
            timeout=timeout,
        )
    except asyncio.TimeoutError:
        logger.warning(
            "等待审计日志后台任务超时（剩余 %d 个未完成）",
            len(_pending_audit_tasks),
        )
# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
def _parse_path(
    path: str,
) -> Tuple[Optional[str], Optional[str], Optional[str]]:
    """
    从路径中提取 (module_id, resource, resource_id)。
    例如：
        /api/v1/auth/users          → ("auth", "users", None)
        /api/v1/auth/users/123      → ("auth", "users", "123")
        /api/v1/auth/login          → ("auth", "login", None)
        /api/v1/modules             → ("module_manager", "modules", None)
        /api/v1/audit-logs          → ("audit_log", "audit-logs", None)
        /api/v1/audit-logs/export   → ("audit_log", "export", None)
    v1.1 修复 P1-1：
        单段路径（如 /api/v1/modules、/api/v1/audit-logs）的 resource
        解析为 path_module，避免 None。
    """
    prefix = "/api/v1/"
    if not path.startswith(prefix):
        return None, None, None
    parts = [p for p in path[len(prefix):].split("/") if p]
    if not parts:
        return None, None, None
    path_module = parts[0]
    module_id = PATH_MODULE_MAP.get(path_module, path_module)
    if len(parts) == 1:
        # P1-1：单段路径，resource 为模块本身
        resource = path_module
        resource_id = None
    else:
        resource = parts[1]
        resource_id = parts[2] if len(parts) > 2 else None
    return module_id, resource, resource_id
def _extract_user(request: Request) -> Tuple[Optional[int], Optional[str]]:
    """
    从 Authorization: Bearer <token> 中解码 JWT，提取 (user_id, username)。
    失败时返回 (None, None)，不抛异常。
    """
    auth_header = request.headers.get("authorization") or ""
    if not auth_header.lower().startswith("bearer "):
        return None, None
    token = auth_header[7:].strip()
    if not token:
        return None, None
    try:
        payload = decode_token(token)
    except Exception:
        return None, None
    user_id_raw = payload.get("sub")
    username = payload.get("username")
    try:
        user_id = int(user_id_raw) if user_id_raw is not None else None
    except (TypeError, ValueError):
        user_id = None
    return user_id, username
def _resolve_action(method: str, path: str) -> str:
    """
    解析操作的 action。
    v1.1 修复 P1-5：
        若路径以 /export 结尾，action="export"，覆盖默认的 HTTP 方法映射。
    """
    if path.endswith("/export"):
        return "export"
    return METHOD_ACTION_MAP.get(method, method.lower())
def _resolve_error_code(response: Response, status_code: int) -> Optional[int]:
    """
    解析错误码。
    v1.1 修复 P1-3：
        优先从响应头 X-Error-Code 读取精确业务错误码；
        不存在则回退到 STATUS_ERROR_MAP 粗略映射。
    """
    header_value = response.headers.get("X-Error-Code")
    if header_value:
        try:
            return int(header_value)
        except (TypeError, ValueError):
            logger.warning("X-Error-Code 响应头格式无效：%s", header_value)
    return STATUS_ERROR_MAP.get(status_code, 90000)
# ---------------------------------------------------------------------------
# 中间件
# ---------------------------------------------------------------------------
class AuditLogMiddleware(BaseHTTPMiddleware):
    """
    审计日志中间件。
    执行顺序（Starlette LIFO）：
        请求 → RequestIdMiddleware → AuditLogMiddleware → CORS → 路由
        响应 ← RequestIdMiddleware ← AuditLogMiddleware ← CORS ← 路由
    """
    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        # 1. 跳过白名单路径
        if any(path.startswith(prefix) for prefix in SKIP_PATH_PREFIXES):
            return await call_next(request)
        # 2. 仅记录 /api/v1/* 请求
        if not path.startswith("/api/v1/"):
            return await call_next(request)
        # 3. 跳过 OPTIONS 预检
        if request.method == "OPTIONS":
            return await call_next(request)
        # 4. 解析路径
        module_id, resource, resource_id = _parse_path(path)
        if module_id is None:
            return await call_next(request)
        # 5. 提取用户
        user_id, username = _extract_user(request)
        # 6. 提取客户端信息与 request_id
        ip = request.client.host if request.client else None
        user_agent = request.headers.get("user-agent")
        request_id = request_id_var.get("unknown")
        # 7. 解析 action（P1-5：/export 特殊处理）
        action = _resolve_action(request.method, path)
        # 8. 调用路由
        response: Optional[Response] = None
        status_code = 500
        status_str = "success"
        error_code: Optional[int] = None
        try:
            response = await call_next(request)
            status_code = response.status_code
        except Exception:
            status_str = "fail"
            error_code = 90000
            raise
        finally:
            if status_code >= 400:
                status_str = "fail"
                # P1-3：优先从 X-Error-Code 读取精确错误码
                if response is not None:
                    error_code = _resolve_error_code(response, status_code)
                else:
                    error_code = STATUS_ERROR_MAP.get(status_code, 90000)
            # 9. P1-2：异步调度写入任务，不阻塞响应
            _schedule_audit_write(
                write_operation_log(
                    user_id=user_id,
                    username=username,
                    module_id=module_id,
                    action=action,
                    resource=resource,
                    resource_id=resource_id,
                    detail=f"{request.method} {path}",
                    ip=ip,
                    user_agent=user_agent,
                    status=status_str,
                    error_code=error_code,
                    request_id=request_id,
                )
            )
        return response
