from typing import Any, Dict, Optional
import logging
from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from src.core.response import error_response
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
                code=10000,  # 通用参数校验错误
                message="请求参数校验失败",
                data={"errors": exc.errors()},
                request_id=request.headers.get("X-Request-ID"),
            ),
        )
    @app.exception_handler(Exception)
    async def generic_exception_handler(request: Request, exc: Exception):
        logger = logging.getLogger(__name__)
        logger.error(f"未捕获异常: {exc}", exc_info=True)
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content=error_response(
                code=99999,
                message="服务器内部错误",
                request_id=request.headers.get("X-Request-ID"),
            ),
        )
