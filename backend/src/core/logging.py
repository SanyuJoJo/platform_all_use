import logging
import sys
from contextvars import ContextVar
from src.core.config import settings
request_id_var: ContextVar[str] = ContextVar("request_id", default="unknown")
class RequestIdFilter(logging.Filter):
    def filter(self, record):
        record.request_id = request_id_var.get()
        return True
def setup_logging():
    handlers = [logging.StreamHandler(sys.stdout)]
    level = getattr(logging, settings.LOG_LEVEL.upper(), logging.INFO)
    logging.basicConfig(
        level=level,
        format="%(asctime)s - %(name)s - %(levelname)s - [%(request_id)s] - %(message)s",
        handlers=handlers,
    )
    for handler in handlers:
        handler.addFilter(RequestIdFilter())
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
    logging.getLogger("sqlalchemy.engine").setLevel(logging.WARNING)
