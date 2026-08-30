from typing import Any, Dict, Optional
from datetime import datetime, timezone
def success_response(
    data: Optional[Any] = None,
    message: str = "success",
    code: int = 0,
    request_id: Optional[str] = None,
) -> Dict[str, Any]:
    """成功响应格式"""
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
    """错误响应格式"""
    return {
        "code": code,
        "message": message,
        "data": data,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "requestId": request_id,
    }
