"""日志审计模块 - 路由定义。
路由前缀：/api/v1/audit-logs
权限点：audit_log:log:view / audit_log:log:export
v1.1 变更：
- P1-4：export_audit_logs 的 format 参数重命名为 export_format，
        通过 Query(alias="format") 保持 URL 兼容（客户端请求 ?format=csv 不变）。
注意：/export 必须在 /{log_id} 之前注册，否则会被路径参数捕获。
"""
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, Depends, Path, Query
from fastapi.responses import Response
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.database import get_db
from src.core.response import success_response
from src.modules.audit_log import service
from src.modules.audit_log.dependencies import CurrentUser, require_permission

router = APIRouter(prefix="/api/v1/audit-logs", tags=["Audit Log"])
@router.get("")
async def list_audit_logs(
    page: int = Query(1, ge=1, description="页码，默认 1"),
    page_size: int = Query(
        20, ge=1, le=100, description="每页条数，默认 20，最大 100"
    ),
    module_id: Optional[str] = Query(None, description="按模块筛选"),
    user_id: Optional[int] = Query(None, description="按用户筛选"),
    action: Optional[str] = Query(None, description="按操作类型筛选"),
    start_time: Optional[datetime] = Query(
        None, description="开始时间（ISO 8601）"
    ),
    end_time: Optional[datetime] = Query(
        None, description="结束时间（ISO 8601）"
    ),
    keyword: Optional[str] = Query(None, description="搜索关键词"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("audit_log:log:view")),
):
    """查询操作日志（分页、多维筛选）。
    - 权限：`audit_log:log:view`
    - 分页格式遵循 API 文档 § 1.3
    - `start_time` / `end_time` 为 ISO 8601 字符串
    - 若 `start_time > end_time`，返回 40001
    """
    data = await service.list_audit_logs(
        db,
        page=page,
        page_size=page_size,
        module_id=module_id,
        user_id=user_id,
        action=action,
        start_time=start_time,
        end_time=end_time,
        keyword=keyword,
    )
    return success_response(data=data)
@router.get("/export")
async def export_audit_logs(
    export_format: str = Query(
        "csv",
        alias="format",
        description="导出格式：csv / json",
    ),
    module_id: Optional[str] = Query(None, description="按模块筛选"),
    user_id: Optional[int] = Query(None, description="按用户筛选"),
    action: Optional[str] = Query(None, description="按操作类型筛选"),
    start_time: Optional[datetime] = Query(None, description="开始时间"),
    end_time: Optional[datetime] = Query(None, description="结束时间"),
    keyword: Optional[str] = Query(None, description="搜索关键词"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("audit_log:log:export")),
):
    """导出操作日志（CSV/JSON）。
    - 权限：`audit_log:log:export`
    - URL Query 参数名保持为 `format`（通过 alias 兼容），
      Python 变量名为 `export_format`（避免与内置 `format` 冲突）
    - `format` 仅支持 `csv` / `json`，否则返回 40004
    - 单次最大导出 10000 行
    - CSV 含 UTF-8 BOM，兼容 Excel 中文显示
    """
    content, media_type, filename = await service.export_audit_logs(
        db,
        export_format=export_format,
        module_id=module_id,
        user_id=user_id,
        action=action,
        start_time=start_time,
        end_time=end_time,
        keyword=keyword,
    )
    return Response(
        content=content,
        media_type=media_type,
        headers={
            "Content-Disposition": f'attachment; filename="{filename}"',
        },
    )
@router.get("/{log_id}")
async def get_audit_log_detail(
    log_id: int = Path(..., gt=0, description="日志 ID"),
    db: AsyncSession = Depends(get_db),
    _: CurrentUser = Depends(require_permission("audit_log:log:view")),
):
    """获取日志详情。
    - 权限：`audit_log:log:view`
    - 日志不存在返回 40002
    """
    data = await service.get_audit_log_detail(db, log_id)
    return success_response(data=data)
