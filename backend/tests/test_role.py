"""
角色管理模块核心接口测试（v1.3）。
运行前请确保：
1. tests/conftest.py 已配置测试数据库（默认 test_app.db）
2. 首次运行时会自动创建表结构与种子数据
v1.3 变更：
- P2-NEW-A：重写 test_update_role_reject_code_field，传入 code 字段，
        断言 422 / 90004；新增 test_update_role_reject_is_system_field，
        覆盖 is_system 字段拒绝。
v1.2 变更：
- P2-NEW-1：新增 temp_role fixture，统一 try/finally 清理。
- P2-NEW-2：test_delete_role_used_by_user 的 finally 块用 try/except 包裹，
        避免清理失败掩盖原 AssertionError。
- 新增 test_map_role_integrity_error_postgresql_format 纯函数测试（P1-NEW-2）。
- 新增 test_create_description_too_long / test_update_description_too_long
        （P1-NEW-3）。
v1.1 变更：
- 覆盖创建 / 更新 / 删除 / 列表 / 权限保护 / 错误码（共 19 个用例）。
- 覆盖 P1-7：permission_codes 去重与格式校验。
- 覆盖 P1-4：所有角色均执行用户使用检查。
- 覆盖 P0-3：系统内置角色基于 is_system 判断。
运行方式：
    uv run pytest tests/test_role.py -v
    KEEP_TEST_DB=1 uv run pytest tests/test_role.py -v   # 调试时保留测试库
"""
import uuid
import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete, select
from sqlalchemy.exc import IntegrityError
from src.core.database import AsyncSessionLocal
from src.main import app
from src.modules.auth.models import Role, UserRole
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
@pytest.fixture
async def temp_role(client: AsyncClient, admin_token: str):
    """P2-NEW-1：创建独立临时角色，测试结束后清理。
    用法：
        async def test_xxx(client, admin_token, temp_role):
            role_id = temp_role["id"]
            ...
    """
    code = f"tr_{uuid.uuid4().hex[:8]}"
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"临时角色_{code}",
            "code": code,
            "description": "临时测试角色",
            "permission_codes": [],
        },
    )
    assert resp.status_code == 200, resp.text
    role_id = resp.json()["data"]["id"]
    try:
        yield {"id": role_id, "code": code}
    finally:
        # 清理失败不影响断言结果
        try:
            await client.delete(
                f"/api/v1/auth/roles/{role_id}",
                headers={"Authorization": f"Bearer {admin_token}"},
            )
        except Exception:
            pass
async def _get_role_id_by_code(code: str):
    """从数据库按编码查询角色 ID。"""
    async with AsyncSessionLocal() as session:
        role = await session.scalar(select(Role).where(Role.code == code))
        return role.id if role else None
# ===========================================================================
# P1-NEW-2：_map_role_integrity_error 纯函数测试
# ===========================================================================
def test_map_role_integrity_error_sqlite_format():
    """v1.2 新增：验证 SQLite 报错格式映射。"""
    from src.modules.auth.role_service import _map_role_integrity_error
    # auth_role.code 冲突 → 20001
    exc = IntegrityError(
        "INSERT INTO auth_role ...",
        {},
        Exception("UNIQUE constraint failed: auth_role.code"),
    )
    assert _map_role_integrity_error(exc).code == 20001
    # 未知约束 → 90003
    exc = IntegrityError(
        "INSERT INTO unknown ...",
        {},
        Exception("UNIQUE constraint failed: unknown.column"),
    )
    assert _map_role_integrity_error(exc).code == 90003
def test_map_role_integrity_error_postgresql_format():
    """v1.2 新增 P1-NEW-2：验证 PostgreSQL 报错格式映射。
    关键：SQLAlchemy 为 auth_role.code 创建的索引名为 ix_auth_role_code，
    报错字符串为 `duplicate key value violates unique constraint
    "ix_auth_role_code"`。
    """
    from src.modules.auth.role_service import _map_role_integrity_error
    # ix_auth_role_code → 20001
    exc = IntegrityError(
        "INSERT INTO auth_role ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "ix_auth_role_code"'
        ),
    )
    assert _map_role_integrity_error(exc).code == 20001
    # uq_auth_role_code（命名约束） → 20001
    exc = IntegrityError(
        "INSERT INTO auth_role ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "uq_auth_role_code"'
        ),
    )
    assert _map_role_integrity_error(exc).code == 20001
    # 未知约束 → 90003
    exc = IntegrityError(
        "INSERT INTO unknown ...",
        {},
        Exception(
            'duplicate key value violates unique constraint "unknown_key"'
        ),
    )
    assert _map_role_integrity_error(exc).code == 90003
# ===========================================================================
# 列表
# ===========================================================================
@pytest.mark.asyncio
async def test_list_roles_success(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/roles?page=1&page_size=10",
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
    # 角色对象应包含 is_system 与 permission_codes（P2-2）
    for item in data["items"]:
        assert "is_system" in item
        assert "permission_codes" in item
@pytest.mark.asyncio
async def test_list_roles_without_token(client):
    """未携带 Token → 401 / 10001。"""
    resp = await client.get("/api/v1/auth/roles")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_list_roles_keyword_filter(client, admin_token):
    """keyword 匹配 code=admin。"""
    resp = await client.get(
        "/api/v1/auth/roles?keyword=admin",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    items = resp.json()["data"]["items"]
    assert any(r["code"] == "admin" for r in items)
# ===========================================================================
# 创建
# ===========================================================================
@pytest.mark.asyncio
async def test_create_role_success(client, admin_token, temp_role):
    """P2-NEW-1：使用 temp_role fixture 自动清理。"""
    resp = await client.get(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["code"] == 0
    assert body["data"]["code"] == temp_role["code"]
    assert body["data"]["is_system"] == 0
@pytest.mark.asyncio
async def test_create_duplicate_code(client, admin_token):
    """重复编码 → 20001。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": "重复编码角色",
            "code": "admin",
            "permission_codes": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 20001
@pytest.mark.asyncio
async def test_create_duplicate_name(client, admin_token):
    """重复名称 → 20004。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": "管理员",
            "code": f"dup_{uuid.uuid4().hex[:8]}",
            "permission_codes": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 20004
@pytest.mark.asyncio
async def test_create_invalid_code_format(client, admin_token):
    """编码格式非法（大写开头）→ 90001。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": "格式错误",
            "code": "BadCode",
            "permission_codes": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_create_invalid_permission_format(client, admin_token):
    """权限编码格式非法 → 20054。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"格式测试_{uuid.uuid4().hex[:6]}",
            "code": f"fmt_{uuid.uuid4().hex[:8]}",
            "permission_codes": ["badformat"],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 20054
@pytest.mark.asyncio
async def test_create_permission_not_found(client, admin_token):
    """权限编码不存在 → 20052。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"权限不存在_{uuid.uuid4().hex[:6]}",
            "code": f"noperm_{uuid.uuid4().hex[:8]}",
            "permission_codes": ["no:such:perm"],
        },
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 20052
@pytest.mark.asyncio
async def test_create_description_too_long(client, admin_token):
    """v1.2 新增 P1-NEW-3：角色描述超长 → 90001。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"描述超长_{uuid.uuid4().hex[:6]}",
            "code": f"desclong_{uuid.uuid4().hex[:8]}",
            "description": "x" * 300,
            "permission_codes": [],
        },
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_create_permission_codes_dedup(client, admin_token, temp_role):
    """P1-7 + P2-NEW-1：permission_codes 保序去重。"""
    resp = await client.put(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "permission_codes": [
                "auth:user:view",
                "auth:user:view",
                "auth:role:view",
                "auth:user:view",
            ]
        },
    )
    assert resp.status_code == 200, resp.text
    permission_codes = resp.json()["data"]["permission_codes"]
    assert permission_codes == ["auth:role:view", "auth:user:view"]
# ===========================================================================
# 详情
# ===========================================================================
@pytest.mark.asyncio
async def test_get_role_detail(client, admin_token):
    role_id = await _get_role_id_by_code("admin")
    assert role_id is not None
    resp = await client.get(
        f"/api/v1/auth/roles/{role_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["code"] == "admin"
    assert data["is_system"] == 1
@pytest.mark.asyncio
async def test_get_role_not_found(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/roles/999999",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 20002
# ===========================================================================
# 更新
# ===========================================================================
@pytest.mark.asyncio
async def test_update_role_permissions_full_replace(client, admin_token, temp_role):
    """P1-7 + P2-NEW-1：permission_codes 全量覆盖。"""
    resp = await client.put(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"permission_codes": ["auth:role:view"]},
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["permission_codes"] == ["auth:role:view"]
@pytest.mark.asyncio
async def test_update_description_too_long(client, admin_token, temp_role):
    """v1.2 新增 P1-NEW-3：更新角色描述超长 → 90001。"""
    resp = await client.put(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"description": "x" * 300},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_update_duplicate_name(client, admin_token):
    """更新为已存在的名称 → 20004。"""
    role_id = await _get_role_id_by_code("guest")
    assert role_id is not None
    resp = await client.put(
        f"/api/v1/auth/roles/{role_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"name": "管理员"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 20004
@pytest.mark.asyncio
async def test_update_role_not_found(client, admin_token):
    resp = await client.put(
        "/api/v1/auth/roles/999999",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"name": "不存在"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 20002
@pytest.mark.asyncio
async def test_update_role_reject_code_field(client, admin_token, temp_role):
    """v1.3 修复 P2-NEW-A：验证 RoleUpdate extra="forbid" 拒绝 code 字段。
    Pydantic 在 RoleUpdate 构造阶段即拒绝未知字段，返回 422 / 90004。
    """
    resp = await client.put(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"code": "new_code"},  # 试图修改不可变字段
    )
    assert resp.status_code == 422, resp.text
    body = resp.json()
    assert body["code"] == 90004
    # 校验错误应包含 body.code 路径
    errors = body["data"]["errors"]
    assert any(
        err.get("loc") == ["body", "code"] and err.get("type") == "extra_forbidden"
        for err in errors
    )
@pytest.mark.asyncio
async def test_update_role_reject_is_system_field(client, admin_token, temp_role):
    """v1.3 新增 P2-NEW-A：验证 RoleUpdate extra="forbid" 拒绝 is_system 字段。"""
    resp = await client.put(
        f"/api/v1/auth/roles/{temp_role['id']}",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={"is_system": 1},  # 试图修改系统内置标志
    )
    assert resp.status_code == 422, resp.text
    body = resp.json()
    assert body["code"] == 90004
    errors = body["data"]["errors"]
    assert any(
        err.get("loc") == ["body", "is_system"]
        and err.get("type") == "extra_forbidden"
        for err in errors
    )
# ===========================================================================
# 删除
# ===========================================================================
@pytest.mark.asyncio
async def test_delete_custom_role_success(client, admin_token):
    """删除自定义角色成功。"""
    create_resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"删除测试_{uuid.uuid4().hex[:6]}",
            "code": f"del_{uuid.uuid4().hex[:8]}",
            "permission_codes": [],
        },
    )
    role_id = create_resp.json()["data"]["id"]
    resp = await client.delete(
        f"/api/v1/auth/roles/{role_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["code"] == 0
    # 再次查询应 404
    resp2 = await client.get(
        f"/api/v1/auth/roles/{role_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp2.status_code == 404
    assert resp2.json()["code"] == 20002
@pytest.mark.asyncio
async def test_delete_system_role_forbidden(client, admin_token):
    """P0-3：系统内置角色（is_system=1）删除 → 20003。"""
    role_id = await _get_role_id_by_code("admin")
    assert role_id is not None
    resp = await client.delete(
        f"/api/v1/auth/roles/{role_id}",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20003
@pytest.mark.asyncio
async def test_delete_role_used_by_user(client, admin_token):
    """P1-4 + P2-NEW-2：角色被用户使用时删除 → 20005（对所有角色生效）。
    v1.2 修复 P2-NEW-2：finally 块用 try/except 包裹清理逻辑，
    避免清理失败掩盖原 AssertionError。
    """
    create_resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {admin_token}"},
        json={
            "name": f"占用测试_{uuid.uuid4().hex[:6]}",
            "code": f"used_{uuid.uuid4().hex[:8]}",
            "permission_codes": [],
        },
    )
    role_id = create_resp.json()["data"]["id"]
    # 获取 admin 用户 ID
    me_resp = await client.get(
        "/api/v1/auth/me",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    admin_id = me_resp.json()["data"]["id"]
    # 建立用户-角色关联
    async with AsyncSessionLocal() as session:
        session.add(UserRole(user_id=admin_id, role_id=role_id))
        await session.commit()
    try:
        resp = await client.delete(
            f"/api/v1/auth/roles/{role_id}",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 409
        assert resp.json()["code"] == 20005
    finally:
        # v1.2 修复 P2-NEW-2：清理失败不影响断言结果
        try:
            async with AsyncSessionLocal() as session:
                await session.execute(
                    delete(UserRole).where(
                        UserRole.user_id == admin_id,
                        UserRole.role_id == role_id,
                    )
                )
                await session.commit()
            await client.delete(
                f"/api/v1/auth/roles/{role_id}",
                headers={"Authorization": f"Bearer {admin_token}"},
            )
        except Exception:
            pass
@pytest.mark.asyncio
async def test_delete_role_not_found(client, admin_token):
    resp = await client.delete(
        "/api/v1/auth/roles/999999",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 20002
# ===========================================================================
# 权限校验
# ===========================================================================
@pytest.mark.asyncio
async def test_guest_cannot_create_role(client, guest_token):
    """guest 缺少 auth:role:create → 20051。"""
    resp = await client.post(
        "/api/v1/auth/roles",
        headers={"Authorization": f"Bearer {guest_token}"},
        json={
            "name": f"无权限_{uuid.uuid4().hex[:6]}",
            "code": f"guest_{uuid.uuid4().hex[:8]}",
            "permission_codes": [],
        },
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
