"""
认证模块核心接口测试。
运行前请确保已执行数据库迁移。
"""
import pytest
from httpx import ASGITransport, AsyncClient
from src.main import app
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
# ---------------------------------------------------------------------------
# 登录
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_login_success(client):
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    assert body["data"]["user"]["username"] == "admin"
    assert "access_token" in body["data"]
    assert "refresh_token" in body["data"]
    assert body["data"]["expires_in"] == 1440
@pytest.mark.asyncio
async def test_login_wrong_password_returns_10001(client):
    """密码错误返回 10001，而不是 422。"""
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "wrong-password"},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_login_short_password_returns_10001(client):
    """短密码不触发 422，返回业务错误码 10001。"""
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "x"},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_login_nonexistent_user(client):
    resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "no_such_user", "password": "123456"},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_login_disabled_user(client, disabled_user):
    """禁用用户登录返回 10002（使用独立测试用户，无顺序依赖）。"""
    resp = await client.post(
        "/api/v1/auth/login",
        json={
            "username": disabled_user["username"],
            "password": disabled_user["password"],
        },
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 10002
# ---------------------------------------------------------------------------
# 当前用户
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_me(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/me",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["username"] == "admin"
    assert "auth:user:view" in data["permissions"]
    assert "admin" in data["roles"]
@pytest.mark.asyncio
async def test_me_without_token(client):
    resp = await client.get("/api/v1/auth/me")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_me_invalid_token(client):
    resp = await client.get(
        "/api/v1/auth/me",
        headers={"Authorization": "Bearer invalid.token.value"},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
# ---------------------------------------------------------------------------
# 刷新 Token
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_refresh_success(client):
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    refresh_token = login_resp.json()["data"]["refresh_token"]
    resp = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": refresh_token},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    assert "access_token" in body["data"]
    assert "refresh_token" in body["data"]
    assert body["data"]["refresh_token"] != refresh_token
@pytest.mark.asyncio
async def test_refresh_rotates_token(client):
    """v1.2：刷新后旧 Refresh Token 立即失效，新 Token 可继续刷新。"""
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    old_refresh = login_resp.json()["data"]["refresh_token"]
    first = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": old_refresh},
    )
    assert first.status_code == 200
    new_refresh = first.json()["data"]["refresh_token"]
    assert new_refresh != old_refresh
    # 旧 Token 复用：应被拒绝
    reuse = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": old_refresh},
    )
    assert reuse.status_code == 401
    assert reuse.json()["code"] == 10001
    # 新 Token 可以继续刷新
    second = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": new_refresh},
    )
    assert second.status_code == 200
    assert second.json()["code"] == 0
@pytest.mark.asyncio
async def test_refresh_invalid_token(client):
    resp = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": "invalid.refresh.token"},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_refresh_access_token_rejected(client):
    """使用 access_token 调用刷新接口应被拒绝。"""
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    access_token = login_resp.json()["data"]["access_token"]
    resp = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": access_token},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
# ---------------------------------------------------------------------------
# 登出
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_logout_revokes_refresh_token(client):
    """登出后，原 refresh_token 无法再刷新。"""
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    data = login_resp.json()["data"]
    access_token = data["access_token"]
    refresh_token = data["refresh_token"]
    logout_resp = await client.post(
        "/api/v1/auth/logout",
        headers={"Authorization": f"Bearer {access_token}"},
        json={"refresh_token": refresh_token},
    )
    assert logout_resp.status_code == 200
    assert logout_resp.json()["code"] == 0
    resp = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": refresh_token},
    )
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_logout_idempotent(client):
    """v1.3（N-6）：重复登出同一 Refresh Token 不报错（幂等）。"""
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": "admin", "password": "123456"},
    )
    data = login_resp.json()["data"]
    access_token = data["access_token"]
    refresh_token = data["refresh_token"]
    first = await client.post(
        "/api/v1/auth/logout",
        headers={"Authorization": f"Bearer {access_token}"},
        json={"refresh_token": refresh_token},
    )
    assert first.status_code == 200
    # 再次登出同一 Token，应仍然返回成功（幂等）
    second = await client.post(
        "/api/v1/auth/logout",
        headers={"Authorization": f"Bearer {access_token}"},
        json={"refresh_token": refresh_token},
    )
    assert second.status_code == 200
    assert second.json()["code"] == 0
# ---------------------------------------------------------------------------
# 修改密码
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_change_password_wrong_old(client, admin_token):
    resp = await client.put(
        "/api/v1/auth/me/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "old_password": "wrong-password",
            "new_password": "654321",
            "confirm_password": "654321",
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10003
@pytest.mark.asyncio
async def test_change_password_mismatch(client, admin_token):
    resp = await client.put(
        "/api/v1/auth/me/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "old_password": "123456",
            "new_password": "654321",
            "confirm_password": "abcdef",
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10004
@pytest.mark.asyncio
async def test_change_password_too_short(client, admin_token):
    """新密码过短，返回 10007（而非 422）。"""
    resp = await client.put(
        "/api/v1/auth/me/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "old_password": "123456",
            "new_password": "x",
            "confirm_password": "x",
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10007
@pytest.mark.asyncio
async def test_change_password_empty_old(client, admin_token):
    """空旧密码返回 10003，而不是 422（P2-2）。"""
    resp = await client.put(
        "/api/v1/auth/me/password",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "old_password": "",
            "new_password": "654321",
            "confirm_password": "654321",
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 10003
@pytest.mark.asyncio
async def test_change_password_success(client, password_test_user):
    """
    修改密码成功，且原 Refresh Token 被撤销（使用独立测试用户，无污染）。
    """
    username = password_test_user["username"]
    old_password = password_test_user["password"]
    login_resp = await client.post(
        "/api/v1/auth/login",
        json={"username": username, "password": old_password},
    )
    assert login_resp.status_code == 200, login_resp.text
    data = login_resp.json()["data"]
    token = data["access_token"]
    refresh_token = data["refresh_token"]
    resp = await client.put(
        "/api/v1/auth/me/password",
        headers={"Authorization": f"Bearer {token}"},
        json={
            "old_password": old_password,
            "new_password": "abcdef",
            "confirm_password": "abcdef",
        },
    )
    assert resp.status_code == 200
    assert resp.json()["code"] == 0
    # 原 Refresh Token 应已被撤销
    refresh_resp = await client.post(
        "/api/v1/auth/refresh",
        json={"refresh_token": refresh_token},
    )
    assert refresh_resp.status_code == 401
    # 用新密码登录
    new_login = await client.post(
        "/api/v1/auth/login",
        json={"username": username, "password": "abcdef"},
    )
    assert new_login.status_code == 200
    assert new_login.json()["code"] == 0
# ---------------------------------------------------------------------------
# 种子数据幂等性（P0-1 / P0-2）
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_seed_data_preserves_admin_permissions():
    """
    多次调用 ensure_auth_seed_data 不会删除 admin 角色的已有权限。
    模拟场景：模块安装时为 admin 额外分配一个业务权限，
    再次执行种子数据后该权限应仍然存在。
    v1.3（N-3）：删除第一次未预加载的冗余查询，只保留 selectinload 版本。
    """
    from sqlalchemy import select
    from sqlalchemy.orm import selectinload
    from src.core.database import AsyncSessionLocal
    from src.modules.auth.models import Permission, Role
    from src.modules.auth.service import ensure_auth_seed_data
    async with AsyncSessionLocal() as session:
        # 模拟模块安装时注册的业务权限
        biz_perm = await session.scalar(
            select(Permission).where(Permission.code == "test_module:test:view")
        )
        if not biz_perm:
            biz_perm = Permission(
                code="test_module:test:view",
                name="测试模块查看",
                module_id="test_module",
                resource="test",
                action="view",
            )
            session.add(biz_perm)
            await session.commit()
            await session.refresh(biz_perm)
        # 将业务权限分配给 admin（使用 selectinload 一次性查询）
        admin_role = await session.scalar(
            select(Role)
            .options(selectinload(Role.permissions))
            .where(Role.code == "admin")
        )
        existing_codes = {p.code for p in admin_role.permissions}
        if biz_perm.code not in existing_codes:
            admin_role.permissions.append(biz_perm)
            await session.commit()
        # 再次运行种子数据
        await ensure_auth_seed_data(session)
        # 校验业务权限仍在
        admin_role = await session.scalar(
            select(Role)
            .options(selectinload(Role.permissions))
            .where(Role.code == "admin")
        )
        codes_after = {p.code for p in admin_role.permissions}
        assert "test_module:test:view" in codes_after, (
            "种子数据不应删除 admin 角色已分配的业务权限"
        )
        # 清理
        admin_role.permissions = [
            p for p in admin_role.permissions if p.code != "test_module:test:view"
        ]
        await session.delete(biz_perm)
        await session.commit()
