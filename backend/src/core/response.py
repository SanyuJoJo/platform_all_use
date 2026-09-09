from typing import Any, Dict, Optional
from datetime import datetime, timezone
from src.core.logging import request_id_var

def success_response(
    data: Optional[Any] = None,
    message: str = "success",
    code: int = 0,
    request_id: Optional[str] = None,
) -> Dict[str, Any]:
    if request_id is None:
        request_id = request_id_var.get("unknown")
    return {
        "code": code,
        "message": message,
        "data": data,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "requestId": request_id,
    }

def error_response(
    code: int,
    message: str,
    data: Optional[Any] = None,
    request_id: Optional[str] = None,
) -> Dict[str, Any]:
    if request_id is None:
        request_id = request_id_var.get("unknown")
    return {
        "code": code,
        "message": message,
        "data": data,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "requestId": request_id,
    }
