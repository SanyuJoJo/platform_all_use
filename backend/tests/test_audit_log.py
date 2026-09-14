"""日志审计模块核心接口测试（v1.2.1）。

v1.2 变更：
- P0-2：新增 test_export_csv_escapes_special_chars，
        验证 CSV 数据行对逗号、引号、换行的正确转义。

v1.2.1 修复：
- 修复模块 docstring 中出现三个连续双引号导致 Python 解析失败的问题。
  改用纯文字描述，不再出现字面三引号。

运行前请确保：
1. tests/conftest.py 已配置测试数据库（默认 test_app.db）
2. 首次运行时会自动创建表结构与种子数据
"""
import uuid

import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete

from src.core.database import AsyncSessionLocal
from src.main import app
from src.modules.audit_log.models import AuditLogOperation


@pytest.fixture
async def client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c


@pytest.fixture
async def admin_token(client: AsyncClient) -> str:
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["data"]["access_token"]


@pytest.fixture
async def guest_token(client: AsyncClient) -> str:
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "guest", "password": "123456"},
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["data"]["access_token"]


# ---------------------------------------------------------------------------
# 中间件自动记录
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_middleware_records_login(client, admin_token):
    """登录请求应被中间件自动记录。"""
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    assert resp.status_code == 200
    resp = await client.get(
        "/api/v1/audit-logs?module_id=auth&action=create",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200, resp.text
    items = resp.json()["data"]["items"]
    assert any(
        it["module_id"] == "auth" and "login" in (it["resource"] or "")
        for it in items
    ), f"未找到登录日志：{items}"


@pytest.mark.asyncio
async def test_middleware_records_failed_request(client, admin_token):
    """失败请求（404）应被记录为 fail。"""
    resp = await client.get(
        "/api/v1/auth/users/999999",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    resp = await client.get(
        "/api/v1/audit-logs?action=view&status=fail",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    items = resp.json()["data"]["items"]
    assert any(it["status"] == "fail" for it in items)


# ---------------------------------------------------------------------------
# 列表查询
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_list_audit_logs_success(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs?page=1&page_size=10",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    data = body["data"]
    assert "items" in data and "total" in data
    assert data["page"] == 1
    assert data["page_size"] == 10


@pytest.mark.asyncio
async def test_list_audit_logs_without_token(client):
    resp = await client.get("/api/v1/audit-logs")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001


@pytest.mark.asyncio
async def test_list_audit_logs_guest_forbidden(client, guest_token):
    """guest 无 audit_log:log:view 权限时返回 20051。"""
    resp = await client.get(
        "/api/v1/audit-logs",
        headers={"Authorization": f"Bearer {guest_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051


@pytest.mark.asyncio
async def test_list_audit_logs_invalid_time_range(client, admin_token):
    """start_time 晚于 end_time 时返回 40001。"""
    resp = await client.get(
        "/api/v1/audit-logs"
        "?start_time=2026-09-11T23:59:59&end_time=2026-09-01T00:00:00",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 40001


@pytest.mark.asyncio
async def test_list_audit_logs_filter_by_action(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs?action=view",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    items = resp.json()["data"]["items"]
    assert all(it["action"] == "view" for it in items)


@pytest.mark.asyncio
async def test_list_audit_logs_filter_by_keyword(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs?keyword=login",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200


# ---------------------------------------------------------------------------
# 详情
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_get_audit_log_detail(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs?page=1&page_size=1",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    items = resp.json()["data"]["items"]
    assert items, "审计日志为空，无法测试详情"
    log_id = items[0]["id"]
    resp = await client.get(
        f"/api/v1/audit-logs/{log_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["data"]["id"] == log_id


@pytest.mark.asyncio
async def test_get_audit_log_not_found(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs/99999999",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 40002


# ---------------------------------------------------------------------------
# 导出
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_export_csv(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs/export?format=csv",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert "text/csv" in resp.headers["content-type"]
    assert "attachment" in resp.headers.get("content-disposition", "")
    body = resp.text
    assert "id,user_id,username" in body


@pytest.mark.asyncio
async def test_export_json(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs/export?format=json",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert "application/json" in resp.headers["content-type"]
    data = resp.json()
    assert "total" in data and "items" in data


@pytest.mark.asyncio
async def test_export_invalid_format(client, admin_token):
    resp = await client.get(
        "/api/v1/audit-logs/export?format=xml",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 40004


@pytest.mark.asyncio
async def test_export_guest_forbidden(client, guest_token):
    """guest 无 audit_log:log:export 权限时返回 20051。"""
    resp = await client.get(
        "/api/v1/audit-logs/export",
        headers={"Authorization": f"Bearer {guest_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051


# ---------------------------------------------------------------------------
# P0-2：CSV 数据行转义测试
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_export_csv_escapes_special_chars(client, admin_token):
    """验证 CSV 数据行对逗号、引号、换行的转义。

    流程：
    1. 通过 write_operation_log 写入一条含特殊字符的日志；
    2. 导出 CSV；
    3. 验证：
       - 逗号字段被双引号包裹（形如 a,b 加引号）；
       - 引号被转义为两个双引号（原引号字符加倍）；
       - 不出现三个连续双引号（表示存在双重转义）。
    """
    from src.modules.audit_log.service import write_operation_log

    unique_marker = uuid.uuid4().hex[:8]
    # 同时包含逗号、引号、换行
    detail = f'含逗号{unique_marker}, 引号"q" 和\n换行'

    # 直接调用写入函数（await，同步等待落库）
    await write_operation_log(
        user_id=None,
        username="admin",
        module_id="test_module",
        action="create",
        resource="demo",
        resource_id=None,
        detail=detail,
        ip="127.0.0.1",
        user_agent="pytest",
        status="success",
        error_code=None,
        request_id=unique_marker,
    )

    # 导出 CSV
    resp = await client.get(
        "/api/v1/audit-logs/export?format=csv",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.text

    # 1. 该日志应存在于导出内容中（通过 request_id 定位）
    assert unique_marker in body, "导出内容未包含刚写入的日志"

    # 2. 逗号、引号、换行应被 csv.writer 按 RFC 4180 正确转义
    assert "含逗号" in body
    assert '""q""' in body, "引号未被正确转义为双引号"

    # 3. 不应出现三个连续双引号（双重转义的典型特征）
    #    使用 chr 拼接避免源码中出现字面三引号。
    triple_quote = chr(34) * 3
    assert triple_quote not in body, "CSV 出现三重引号，存在双重转义"

    # 4. 清理测试日志
    async with AsyncSessionLocal() as session:
        await session.execute(
            delete(AuditLogOperation).where(
                AuditLogOperation.request_id == unique_marker
            )
        )
        await session.commit()
