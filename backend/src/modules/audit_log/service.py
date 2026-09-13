"""日志审计模块 - 业务逻辑。

v1.1 变更：
- P0-1：_to_csv 删除 _csv_escape 调用，全部交由 csv.writer 按 RFC 4180
        转义。删除 _csv_escape 函数。
- P1-4：export_audit_logs 的 format 参数重命名为 export_format。
- P2-1：CSV 表头与数据行统一由 csv.writer 处理。

包含：
- write_operation_log：中间件调用的写入函数（独立会话）
- list_audit_logs：分页查询
- get_audit_log_detail：详情
- export_audit_logs：导出 CSV / JSON
"""
import csv
import io
import json
import logging
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional, Tuple

from sqlalchemy import func, or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.database import AsyncSessionLocal
from src.core.exceptions import PlatformException
from src.modules.audit_log.constants import (
    CSV_BOM,
    EXPORT_FORMATS,
    EXPORT_MAX_ROWS,
)
from src.modules.audit_log.models import AuditLogOperation

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# 内部工具
# ---------------------------------------------------------------------------
def _utcnow_naive() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


def _normalize_datetime(dt: Optional[datetime]) -> Optional[datetime]:
    """将带时区的 datetime 转为 UTC naive，与数据库存储格式一致。"""
    if dt is None:
        return None
    if dt.tzinfo is not None:
        return dt.astimezone(timezone.utc).replace(tzinfo=None)
    return dt


def _serialize_log(log: AuditLogOperation) -> Dict[str, Any]:
    """将 AuditLogOperation ORM 对象序列化为统一字典。"""
    return {
        "id": log.id,
        "user_id": log.user_id,
        "username": log.username,
        "module_id": log.module_id,
        "action": log.action,
        "resource": log.resource,
        "resource_id": log.resource_id,
        "detail": log.detail,
        "ip": log.ip,
        "user_agent": log.user_agent,
        "status": log.status,
        "error_code": log.error_code,
        "request_id": log.request_id,
        "created_at": log.created_at,
    }


def _build_filters(
    *,
    module_id: Optional[str],
    user_id: Optional[int],
    action: Optional[str],
    start_time: Optional[datetime],
    end_time: Optional[datetime],
    keyword: Optional[str],
) -> List[Any]:
    """构造查询条件列表（list_audit_logs 与 export_audit_logs 共用）。"""
    conditions: List[Any] = []
    if module_id:
        conditions.append(AuditLogOperation.module_id == module_id)
    if user_id is not None:
        conditions.append(AuditLogOperation.user_id == user_id)
    if action:
        conditions.append(AuditLogOperation.action == action)
    if start_time:
        conditions.append(AuditLogOperation.created_at >= start_time)
    if end_time:
        conditions.append(AuditLogOperation.created_at <= end_time)
    if keyword:
        like = f"%{keyword}%"
        conditions.append(
            or_(
                AuditLogOperation.detail.ilike(like),
                AuditLogOperation.username.ilike(like),
                AuditLogOperation.resource.ilike(like),
            )
        )
    return conditions


# ---------------------------------------------------------------------------
# 写入（中间件调用）
# ---------------------------------------------------------------------------
async def write_operation_log(
    *,
    user_id: Optional[int] = None,
    username: Optional[str] = None,
    module_id: str,
    action: str,
    resource: Optional[str] = None,
    resource_id: Optional[str] = None,
    detail: Optional[str] = None,
    ip: Optional[str] = None,
    user_agent: Optional[str] = None,
    status: str = "success",
    error_code: Optional[int] = None,
    request_id: Optional[str] = None,
) -> None:
    """
    写入一条操作日志。

    - 使用独立数据库会话（AsyncSessionLocal），不依赖请求会话。
    - 内部捕获所有异常并记录 ERROR 日志，绝不影响主请求流程。
    """
    try:
        async with AsyncSessionLocal() as session:
            log = AuditLogOperation(
                user_id=user_id,
                username=username,
                module_id=module_id,
                action=action,
                resource=resource,
                resource_id=resource_id,
                detail=detail,
                ip=ip,
                user_agent=(user_agent or "")[:255] or None,
                status=status,
                error_code=error_code,
                request_id=request_id,
            )
            session.add(log)
            await session.commit()
    except Exception as exc:
        # 审计日志写入失败绝不能影响主流程
        logger.error("写入审计日志失败：%s", exc)


# ---------------------------------------------------------------------------
# 查询
# ---------------------------------------------------------------------------
async def list_audit_logs(
    db: AsyncSession,
    *,
    page: int,
    page_size: int,
    module_id: Optional[str] = None,
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    keyword: Optional[str] = None,
) -> Dict[str, Any]:
    """分页查询操作日志。"""
    start_time = _normalize_datetime(start_time)
    end_time = _normalize_datetime(end_time)

    if start_time and end_time and start_time > end_time:
        raise PlatformException(
            code=40001,
            message="start_time 不能晚于 end_time",
            status_code=400,
        )

    conditions = _build_filters(
        module_id=module_id,
        user_id=user_id,
        action=action,
        start_time=start_time,
        end_time=end_time,
        keyword=keyword,
    )

    count_stmt = select(func.count(AuditLogOperation.id))
    if conditions:
        count_stmt = count_stmt.where(*conditions)
    total = (await db.execute(count_stmt)).scalar_one()

    list_stmt = select(AuditLogOperation)
    if conditions:
        list_stmt = list_stmt.where(*conditions)
    list_stmt = (
        list_stmt.order_by(
            AuditLogOperation.created_at.desc(),
            AuditLogOperation.id.desc(),
        )
        .offset((page - 1) * page_size)
        .limit(page_size)
    )
    result = await db.execute(list_stmt)
    logs = list(result.scalars().all())

    return {
        "items": [_serialize_log(log) for log in logs],
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }


async def get_audit_log_detail(
    db: AsyncSession, log_id: int
) -> Dict[str, Any]:
    """获取日志详情，不存在返回 40002。"""
    log = await db.get(AuditLogOperation, log_id)
    if not log:
        raise PlatformException(
            code=40002, message="日志不存在", status_code=404
        )
    return _serialize_log(log)


# ---------------------------------------------------------------------------
# 导出
# ---------------------------------------------------------------------------
async def export_audit_logs(
    db: AsyncSession,
    *,
    export_format: str,
    module_id: Optional[str] = None,
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    keyword: Optional[str] = None,
) -> Tuple[str, str, str]:
    """
    导出操作日志。

    参数：
        export_format: 导出格式，仅支持 "csv" / "json"（v1.1 P1-4 重命名）

    返回：
        (content, media_type, filename)
    """
    if export_format not in EXPORT_FORMATS:
        raise PlatformException(
            code=40004,
            message=f"导出格式不支持：{export_format}，仅支持 {'/'.join(EXPORT_FORMATS)}",
            status_code=400,
        )

    start_time = _normalize_datetime(start_time)
    end_time = _normalize_datetime(end_time)
    if start_time and end_time and start_time > end_time:
        raise PlatformException(
            code=40001,
            message="start_time 不能晚于 end_time",
            status_code=400,
        )

    conditions = _build_filters(
        module_id=module_id,
        user_id=user_id,
        action=action,
        start_time=start_time,
        end_time=end_time,
        keyword=keyword,
    )

    stmt = select(AuditLogOperation)
    if conditions:
        stmt = stmt.where(*conditions)
    stmt = (
        stmt.order_by(
            AuditLogOperation.created_at.desc(),
            AuditLogOperation.id.desc(),
        )
        .limit(EXPORT_MAX_ROWS)
    )
    result = await db.execute(stmt)
    logs = list(result.scalars().all())

    timestamp = _utcnow_naive().strftime("%Y%m%d_%H%M%S")

    try:
        if export_format == "csv":
            content = _to_csv(logs)
            media_type = "text/csv; charset=utf-8"
            filename = f"audit_logs_{timestamp}.csv"
        else:  # json
            content = _to_json(logs)
            media_type = "application/json; charset=utf-8"
            filename = f"audit_logs_{timestamp}.json"
    except Exception as exc:
        raise PlatformException(
            code=40005,
            message=f"日志导出失败：{exc}",
            status_code=500,
        ) from exc

    return content, media_type, filename


def _to_csv(logs: List[AuditLogOperation]) -> str:
    """
    生成 CSV 内容（含 UTF-8 BOM）。

    v1.1 修复 P0-1：
        - 不再手动调用 _csv_escape 转义字段；
        - 全部交由 csv.writer 按 RFC 4180 处理；
        - 避免双重转义：三层引号会破坏数据，正确形式应为单层双引号包裹。

    v1.1 修复 P2-1：
        - 表头与数据行统一由 csv.writer 处理，风格一致。

    示例（RFC 4180 转义规则由 csv.writer 实现）：
        输入 a,b           → 输出 "a,b"
        输入 a"b           → 输出 "a""b"
        输入 a 换行 b      → 输出 "a\nb"（字段被双引号包裹）
    """
    buf = io.StringIO()
    buf.write(CSV_BOM)
    writer = csv.writer(buf, lineterminator="\n")
    # 表头：同样交由 csv.writer 处理
    writer.writerow(
        [
            "id",
            "user_id",
            "username",
            "module_id",
            "action",
            "resource",
            "resource_id",
            "detail",
            "ip",
            "user_agent",
            "status",
            "error_code",
            "request_id",
            "created_at",
        ]
    )
    for log in logs:
        # P0-1：直接传原始值，由 csv.writer 统一按 RFC 4180 转义
        writer.writerow(
            [
                log.id,
                log.user_id if log.user_id is not None else "",
                log.username if log.username is not None else "",
                log.module_id if log.module_id is not None else "",
                log.action if log.action is not None else "",
                log.resource if log.resource is not None else "",
                log.resource_id if log.resource_id is not None else "",
                log.detail if log.detail is not None else "",
                log.ip if log.ip is not None else "",
                log.user_agent if log.user_agent is not None else "",
                log.status if log.status is not None else "",
                log.error_code if log.error_code is not None else "",
                log.request_id if log.request_id is not None else "",
                log.created_at.isoformat() if log.created_at else "",
            ]
        )
    return buf.getvalue()


def _to_json(logs: List[AuditLogOperation]) -> str:
    """生成 JSON 内容。"""
    items = [_serialize_log(log) for log in logs]
    for item in items:
        if item.get("created_at") is not None:
            item["created_at"] = item["created_at"].isoformat()
    return json.dumps(
        {"total": len(items), "items": items},
        ensure_ascii=False,
        indent=2,
    )
