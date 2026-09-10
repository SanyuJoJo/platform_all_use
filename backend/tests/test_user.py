"""
用户管理模块核心接口测试（v1.3）
v1.3 变更：
- P2-3：新增 test_map_integrity_error_sqlite_format /
        test_map_integrity_error_postgresql_format 纯函数测试用例，
        验证 _map_integrity_error 对两种数据库报错格式的映射。
v1.2 引入：
- P2-4：新增 test_create_user_nickname_too_long /
        test_update_user_nickname_too_long
v1.1 引入：
- P1-3：动态获取 admin id / guest role id
- P2-1：test_list_users_without_token
- P2-4：temp_user fixture 加 try/finally
v1.0 引入：
- 全部核心接口测试
运行前请确保：
1. tests/conftest.py 已配置测试数据库（默认 test_app.db）
2. 首次运行时会自动创建表结构与种子数据
"""
import uuid
from typing import Optional
import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete, select
from sqlalchemy.exc import IntegrityError
from src.core.database import AsyncSessionLocal
from src.core.security import hash_password
from src.main import app
from src.modules.auth.models import Role, User, UserRole
# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------
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
async def _get_current_user_id(client: AsyncClient, token: str) -> int:
    """从 /auth/me 动态获取当前登录用户 ID。"""
    resp = await client.get(
        "/api/v1/auth/me",
        headers={"Authorization": f"Bearer {token}"},
    )
    assert resp.status_code == 200, resp.text
    return int(resp.json()["data"]["id"])
async def _get_role_id_by_code(code: str) -> Optional[int]:
    """从数据库查询角色 ID。"""
    async with AsyncSessionLocal() as session:
        role = await session.scalar(select(Role).where(Role.code == code))
        return role.id if role else None
@pytest.fixture
async def guest_role_id() -> int:
    """动态获取 guest 角色 ID。"""
    role_id = await _get_role_id_by_code("guest")
    if role_id is None:
        pytest.skip("guest 角色不存在，请先初始化种子数据")
    return role_id
@pytest.fixture
async def temp_user():
    """创建独立临时用户，测试结束后清理。"""
    username = f"tu_{uuid.uuid4().hex[:8]}"
    async with AsyncSessionLocal() as session:
        user = User(
            username=username,
            password_hash=hash_password("123456"),
            nickname="临时用户",
            email=f"{username}@example.com",
            status=1,
        )
        session.add(user)
        await session.commit()
        await session.refresh(user)
        user_id = user.id
    try:
        yield {"id": user_id, "username": username, "password": "123456"}
    finally:
        async with AsyncSessionLocal() as session:
            await session.execute(delete(UserRole).where(UserRole.user_id == user_id))
            user = await session.get(User, user_id)
            if user:
                await session.delete(user)
                await session.commit()
# ---------------------------------------------------------------------------
# P2-3：_map_integrity_error 纯函数测试
# ---------------------------------------------------------------------------
def test_map_integrity_error_sqlite_format():
    """
    P2-3：验证 _map_integrity_error 对 SQLite 报错格式的映射。
    SQLite 报错格式：
        UNIQUE constraint failed: auth_user.username
        UNIQUE constraint failed: auth_user.email
        UNIQUE constraint failed: auth_user_role.user_id, auth_user_role.role_id
    """
    from src.modules.auth.user_service import _map_integrity_error
    # username 冲突 → 10000
    exc = IntegrityError(
        "INSERT INTO auth_user ...",
        {},
        Exception("UNIQUE constraint failed: auth_user.username"),
    )
    assert _map_integrity_error(exc).code == 10000
    # email 冲突 → 10009
    exc = IntegrityError(
        "INSERT INTO auth_user ...",
        {},
        Exception("UNIQUE constraint failed: auth_user.email"),
    )
    assert _map_integrity_error(exc).code == 10009
    # user_role 冲突 → 90003
    exc = IntegrityError(
        "INSERT INTO auth_user_role ...",
        {},
        Exception(
            "UNIQUE constraint failed: auth_user_role.user_id, auth_user_role.role_id"
        ),
    )
    assert _map_integrity_error(exc).code == 90003
    # 未知约束 → 90003
    exc = IntegrityError(
        "INSERT INTO unknown ...",
        {},
        Exception("UNIQUE constraint failed: unknown.column"),
    )
    assert _map_integrity_error(exc).code == 90003
def test_map_integrity_error_postgresql_format():
    """
    P2-3：验证 _map_integrity_error 对 PostgreSQL 报错格式的映射。
    PostgreSQL 报错格式：
        duplicate key value violates unique constraint "auth_user_username_key"
        duplicate key value violates unique constraint "auth_user_email_key"
        duplicate key value violates unique constraint "uq_auth_user_role_user_role"
    """
    from src.modules.auth.user_service import _map_integrity_error
    # username 冲突 → 10000
    exc = IntegrityError(
        "INSERT INTO auth_user ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "auth_user_username_key"'
        ),
    )
    assert _map_integrity_error(exc).code == 10000
    # email 冲突 → 10009
    exc = IntegrityError(
        "INSERT INTO auth_user ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "auth_user_email_key"'
        ),
    )
    assert _map_integrity_error(exc).code == 10009
    # user_role 冲突 → 90003
    exc = IntegrityError(
        "INSERT INTO auth_user_role ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "uq_auth_user_role_user_role"'
        ),
    )
    assert _map_integrity_error(exc).code == 90003
    # 未知约束 → 90003
    exc = IntegrityError(
        "INSERT INTO unknown ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "unknown_key"'
        ),
    )
    assert _map_integrity_error(exc).code == 90003
# ---------------------------------------------------------------------------
# 列表
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_list_users_success(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/users?page=1&page_size=10",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    data = body["data"]
    assert "items" in data and "total" in data
    assert data["page"] == 1
    assert data["page_size"] == 10
    assert data["total"] >= 2
@pytest.mark.asyncio
async def test_list_users_without_token(client):
    """未携带 Token 时返回 10001。"""
    resp = await client.get("/api/v1/auth/users")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_list_users_keyword_filter(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/users?keyword=admin",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    items = resp.json()["data"]["items"]
    assert any(u["username"] == "admin" for u in items)
# ---------------------------------------------------------------------------
# 创建
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_create_user_success(client, admin_token, guest_role_id):
    username = f"new_{uuid.uuid4().hex[:8]}"
    try:
        resp = await client.post(
            "/api/v1/auth/users",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={
                "username": username,
                "password": "123456",
                "nickname": "新用户",
                "email": f"{username}@example.com",
                "role_ids": [guest_role_id],
            },
        )
        assert resp.status_code == 200, resp.text
        body = resp.json()
        assert body["code"] == 0
        assert body["data"]["username"] == username
        assert body["data"]["roles"][0]["code"] == "guest"
        assert body["data"]["updated_at"] is not None
    finally:
        async with AsyncSessionLocal() as session:
            user = await session.scalar(select(User).where(User.username == username))
            if user:
                await session.execute(
                    delete(UserRole).where(UserRole.user_id == user.id)
                )
                await session.delete(user)
                await session.commit()
@pytest.mark.asyncio
async def test_create_user_duplicate_username(client, admin_token):
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": "admin",
            "password": "123456",
            "nickname": "重复",
            "role_ids": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10000
@pytest.mark.asyncio
async def test_create_user_duplicate_email(client, admin_token):
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": f"dup_{uuid.uuid4().hex[:6]}",
            "password": "123456",
            "nickname": "重复邮箱",
            "email": "admin@example.com",
            "role_ids": [],
        },
    )
    assert resp.status_code == 409
    assert resp.json()["code"] == 10009
@pytest.mark.asyncio
async def test_create_user_invalid_username(client, admin_token):
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": "ab",
            "password": "123456",
            "nickname": "短名",
            "role_ids": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10006
@pytest.mark.asyncio
async def test_create_user_too_long_username(client, admin_token):
    """>20 位用户名也返回 10006（而非 422）。"""
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": "a" * 30,
            "password": "123456",
            "nickname": "长名",
            "role_ids": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10006
@pytest.mark.asyncio
async def test_create_user_nickname_too_long(client, admin_token):
    """昵称超长返回 90001（而非 422）。"""
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": f"nn_{uuid.uuid4().hex[:6]}",
            "password": "123456",
            "nickname": "x" * 100,
            "role_ids": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_create_user_short_password(client, admin_token):
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "username": f"sp_{uuid.uuid4().hex[:6]}",
            "password": "x",
            "nickname": "短密码",
            "role_ids": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10007
# ---------------------------------------------------------------------------
# 详情
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_get_user_detail(client, admin_token, temp_user):
    resp = await client.get(
        f"/api/v1/auth/users/{temp_user['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["data"]["username"] == temp_user["username"]
@pytest.mark.asyncio
async def test_get_user_not_found(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/users/999999",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 10005
# ---------------------------------------------------------------------------
# 更新
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_update_user_nickname_and_roles(client, admin_token, temp_user, guest_role_id):
    resp = await client.put(
        f"/api/v1/auth/users/{temp_user['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"nickname": "已更新", "role_ids": [guest_role_id]},
    )
    assert resp.status_code == 200, resp.text
    data = resp.json()["data"]
    assert data["nickname"] == "已更新"
    assert any(r["code"] == "guest" for r in data["roles"])
@pytest.mark.asyncio
async def test_update_user_nickname_too_long(client, admin_token, temp_user):
    """更新昵称超长返回 90001。"""
    resp = await client.put(
        f"/api/v1/auth/users/{temp_user['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"nickname": "x" * 100},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_update_user_disable_self(client, admin_token):
    """更新自己 status=0 → 10011。"""
    admin_id = await _get_current_user_id(client, admin_token)
    resp = await client.put(
        f"/api/v1/auth/users/{admin_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"status": 0},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 10011
# ---------------------------------------------------------------------------
# 删除
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_delete_user_success(client, admin_token, temp_user):
    resp = await client.delete(
        f"/api/v1/auth/users/{temp_user['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["code"] == 0
    resp2 = await client.get(
        f"/api/v1/auth/users/{temp_user['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp2.status_code == 404
@pytest.mark.asyncio
async def test_delete_self_forbidden(client, admin_token):
    """删除自己 → 10010。"""
    admin_id = await _get_current_user_id(client, admin_token)
    resp = await client.delete(
        f"/api/v1/auth/users/{admin_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 10010
# ---------------------------------------------------------------------------
# 启用/禁用
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_disable_and_enable_user(client, admin_token, temp_user):
    resp = await client.patch(
        f"/api/v1/auth/users/{temp_user['id']}/status",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"status": 0},
    )
    assert resp.status_code == 200
    assert resp.json()["data"]["status"] == 0
    resp2 = await client.patch(
        f"/api/v1/auth/users/{temp_user['id']}/status",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"status": 1},
    )
    assert resp2.status_code == 200
    assert resp2.json()["data"]["status"] == 1
@pytest.mark.asyncio
async def test_disable_self_forbidden(client, admin_token):
    """禁用自己 → 10011。"""
    admin_id = await _get_current_user_id(client, admin_token)
    resp = await client.patch(
        f"/api/v1/auth/users/{admin_id}/status",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"status": 0},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 10011
# ---------------------------------------------------------------------------
# 重置密码
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_reset_password_success(client, admin_token, temp_user):
    resp = await client.patch(
        f"/api/v1/auth/users/{temp_user['id']}/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"new_password": "abcdef"},
    )
    assert resp.status_code == 200
    assert resp.json()["code"] == 0
    login = await client.post(
        "/api/v1/auth/login",
        json={"username": temp_user["username"], "password": "abcdef"},
    )
    assert login.status_code == 200
@pytest.mark.asyncio
async def test_reset_password_too_short(client, admin_token, temp_user):
    resp = await client.patch(
        f"/api/v1/auth/users/{temp_user['id']}/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"new_password": "x"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10007
# ---------------------------------------------------------------------------
# 权限校验
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_guest_cannot_create_user(client, guest_token):
    """guest 缺少 auth:user:create → 20051。"""
    resp = await client.post(
        "/api/v1/auth/users",
        headers={"Authorization": f"Bearer {guest_token}"},
        json={
            "username": f"g_{uuid.uuid4().hex[:6]}",
            "password": "123456",
            "nickname": "无权限",
            "role_ids": [],
        },
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
