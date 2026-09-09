from fastapi import APIRouter, Depends, Query, Path
from typing import Optional
from fastapi.responses import Response
from src.core.dependencies import get_current_user, require_permission
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .service import list_logs, get_log_by_id, export_logs_csv

router = APIRouter(prefix="/api/v1/audit-logs", tags=["Audit Log"])

@router.get("/")
async def get_audit_logs(
    page: int = Query(1, ge=1),
    page_size: int = Query(10, ge=1, le=100),
    module_id: Optional[str] = None,
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    start_time: Optional[str] = None,
    end_time: Optional[str] = None,
    keyword: Optional[str] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("audit_log:log:view"))
):
    result = list_logs(page, page_size, module_id, user_id, action, start_time, end_time, keyword)
    return success_response(data=result)

# ✅ 将导出接口放在 /{log_id} 之前，避免被解析为 log_id
@router.get("/export")
async def export_logs(
    module_id: Optional[str] = None,
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    start_time: Optional[str] = None,
    end_time: Optional[str] = None,
    keyword: Optional[str] = None,
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("audit_log:log:export"))
):
    csv_data = export_logs_csv({
        "module_id": module_id,
        "user_id": user_id,
        "action": action,
        "start_time": start_time,
        "end_time": end_time,
        "keyword": keyword,
    })
    return Response(
        content=csv_data,
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=audit_logs.csv"}
    )

@router.get("/{log_id}")
async def get_log_detail(
    log_id: int = Path(..., gt=0),
    current_user: dict = Depends(get_current_user),
    _ = Depends(require_permission("audit_log:log:view"))
):
    log = get_log_by_id(log_id)
    if not log:
        raise PlatformException(code=40002, message="日志不存在", status_code=404)
    return success_response(data=log)
