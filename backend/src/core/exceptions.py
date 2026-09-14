"""全局异常处理。
v1.1 变更：
- P1-3：所有异常处理器通过 `_error_json_response` 设置 `X-Error-Code`
        响应头，供 AuditLogMiddleware 读取精确业务错误码。
"""
import logging
from typing import Any, Dict, Optional

from fastapi import Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from src.core.response import error_response

logger = logging.getLogger(__name__)
class PlatformException(Exception):
    """自定义业务异常。"""
    def __init__(
        self,
        code: int,
        message: str = "业务异常",
        data: Optional[Dict[str, Any]] = None,
        status_code: int = status.HTTP_400_BAD_REQUEST,
    ):
        self.code = code
        self.message = message
        self.data = data
        self.status_code = status_code
        super().__init__(message)
def _error_json_response(
    *,
    status_code: int,
    code: int,
    message: str,
    data: Optional[Any] = None,
    request_id: Optional[str] = None,
) -> JSONResponse:
    """
    构造统一错误响应，并设置 `X-Error-Code` 响应头。
    - 响应体遵循统一格式 `{code, message, data, timestamp, requestId}`；
    - `X-Error-Code` 供 AuditLogMiddleware 读取精确业务错误码；
    - 若 request_id 为 None，由 `error_response` 从上下文变量中获取。
    """
    response = JSONResponse(
        status_code=status_code,
        content=error_response(
            code=code,
            message=message,
            data=data,
            request_id=request_id,
        ),
    )
    response.headers["X-Error-Code"] = str(code)
    return response
def setup_exception_handlers(app):
    """注册全局异常处理器。"""
    @app.exception_handler(PlatformException)
    async def platform_exception_handler(request: Request, exc: PlatformException):
        return _error_json_response(
            status_code=exc.status_code,
            code=exc.code,
            message=exc.message,
            data=exc.data,
            request_id=request.headers.get("X-Request-ID"),
        )
    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(
        request: Request, exc: RequestValidationError
    ):
        return _error_json_response(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            code=90004,
            message="请求参数校验失败",
            data={"errors": exc.errors()},
            request_id=request.headers.get("X-Request-ID"),
        )
    @app.exception_handler(StarletteHTTPException)
    async def starlette_http_exception_handler(
        request: Request, exc: StarletteHTTPException
    ):
        logger.info(
            "捕获 StarletteHTTPException: status=%s, detail=%s",
            exc.status_code,
            exc.detail,
        )
        if exc.status_code == status.HTTP_404_NOT_FOUND:
            code = 90002
            message = "资源不存在"
        elif exc.status_code == status.HTTP_403_FORBIDDEN:
            code = 20051
            message = "禁止访问"
        elif exc.status_code == status.HTTP_401_UNAUTHORIZED:
            code = 10001
            message = "未认证"
        else:
            code = 90001
            message = exc.detail or "请求错误"
        return _error_json_response(
            status_code=exc.status_code,
            code=code,
            message=message,
            data=None,
            request_id=request.headers.get("X-Request-ID"),
        )
    @app.exception_handler(Exception)
    async def generic_exception_handler(request: Request, exc: Exception):
        logger.error("未捕获异常: %s", exc, exc_info=True)
        return _error_json_response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            code=99999,
            message="服务器内部错误",
            data=None,
            request_id=request.headers.get("X-Request-ID"),
        )
