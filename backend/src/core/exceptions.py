import logging
from typing import Any, Dict, Optional
from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException  # 关键导入
from src.core.response import error_response

logger = logging.getLogger(__name__)

class PlatformException(Exception):
    """自定义业务异常"""
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

def setup_exception_handlers(app):
    """注册全局异常处理器"""

    @app.exception_handler(PlatformException)
    async def platform_exception_handler(request: Request, exc: PlatformException):
        return JSONResponse(
            status_code=exc.status_code,
            content=error_response(
                code=exc.code,
                message=exc.message,
                data=exc.data,
                request_id=request.headers.get("X-Request-ID"),
            ),
        )

    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(request: Request, exc: RequestValidationError):
        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            content=error_response(
                code=90004,
                message="请求参数校验失败",
                data={"errors": exc.errors()},
                request_id=request.headers.get("X-Request-ID"),
            ),
        )

    # 关键：使用 StarletteHTTPException 捕获所有 HTTP 异常（包括 404）
    @app.exception_handler(StarletteHTTPException)
    async def starlette_http_exception_handler(request: Request, exc: StarletteHTTPException):
        logger.info(f"捕获 StarletteHTTPException: status={exc.status_code}, detail={exc.detail}")
        # 映射业务错误码
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
        return JSONResponse(
            status_code=exc.status_code,
            content=error_response(
                code=code,
                message=message,
                data=None,
                request_id=request.headers.get("X-Request-ID"),
            ),
            media_type="application/json",
        )

    @app.exception_handler(Exception)
    async def generic_exception_handler(request: Request, exc: Exception):
        logger.error(f"未捕获异常: {exc}", exc_info=True)
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content=error_response(
                code=99999,
                message="服务器内部错误",
                request_id=request.headers.get("X-Request-ID"),
            ),
        )
