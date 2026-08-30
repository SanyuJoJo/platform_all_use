from uuid import uuid4
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
from src.core.logging import request_id_var
class RequestIdMiddleware(BaseHTTPMiddleware):
    """为每个请求生成或提取 request_id，并注入到日志上下文"""
    async def dispatch(self, request: Request, call_next):
        # 从请求头获取 X-Request-ID，若不存在则生成
        request_id = request.headers.get("X-Request-ID")
        if not request_id:
            request_id = str(uuid4())
        # 设置到上下文变量
        token = request_id_var.set(request_id)
        try:
            # 将 request_id 添加到响应头
            response = await call_next(request)
            response.headers["X-Request-ID"] = request_id
            return response
        finally:
            request_id_var.reset(token)
