import logging
import sys
from contextvars import ContextVar
request_id_var: ContextVar[str] = ContextVar("request_id", default="unknown")
class RequestIdFilter(logging.Filter):
    """为每条日志添加 request_id"""
    def filter(self, record):
        record.request_id = request_id_var.get()
        return True
def setup_logging():
    """配置结构化日志（开发环境可简单输出）"""
    handlers = [logging.StreamHandler(sys.stdout)]
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(name)s - %(levelname)s - [%(request_id)s] - %(message)s",
        handlers=handlers,
    )
    # 添加 request_id 过滤器
    for handler in handlers:
        handler.addFilter(RequestIdFilter())
    # 设置第三方库日志级别
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
    logging.getLogger("sqlalchemy.engine").setLevel(logging.WARNING)
