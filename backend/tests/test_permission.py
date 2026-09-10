"""
权限管理模块核心接口测试（v1.2）。
v1.2 变更：
- P1-NEW-1：重写 test_register_permissions_cross_module_conflict，
        采用"脏数据构造"方案真实触发 20053 分支。
- P1-NEW-2：重写 test_register_permissions_partial_rollback，
        真实验证"部分新增 + 冲突 → 整体回滚"。
- P2-NEW-1：新增 test_list_permissions_invalid_module_id，
        验证查询接口非法 module_id 返回 90001。
- 测试总数：v1.0 的 8 个 + v1.1 新增 9 个 + v1.2 新增 1 个 = 18 个。
运行前请确保：
1. tests/conftest.py 已配置测试数据库（默认 test_app.db）
2. 首次运行时会自动创建表结构与种子数据
"""
import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete, select
from src.core.database import AsyncSessionLocal
from src.core.exceptions import PlatformException
from src.main import app
from src.modules.auth.models import Permission
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
# ===========================================================================
# 查询权限列表
# ===========================================================================
@pytest.mark.asyncio
async def test_list_permissions_success(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/permissions",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["code"] == 0
    data = body["data"]
    assert isinstance(data, list)
    assert len(data) > 0
    codes = {item["code"] for item in data}
    assert "auth:user:view" in codes
@pytest.mark.asyncio
async def test_list_permissions_by_module(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/permissions?module_id=auth",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    for item in body["data"]:
        assert item["module_id"] == "auth"
@pytest.mark.asyncio
async def test_list_permissions_by_resource(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/permissions?module_id=auth&resource=user",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    for item in body["data"]:
        assert item["module_id"] == "auth"
        assert item["resource"] == "user"
@pytest.mark.asyncio
async def test_list_permissions_invalid_module_id(client, admin_token):
    """v1.2 P2-NEW-1：非法 module_id 返回 90001，与 get_permissions_by_module 对齐。"""
    resp = await client.get(
        "/api/v1/auth/permissions?module_id=BadModule",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
@pytest.mark.asyncio
async def test_get_module_permissions(client, admin_token):
    resp = await client.get(
        "/api/v1/auth/permissions/modules/auth",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    assert isinstance(body["data"], list)
    assert any(item["code"] == "auth:role:view" for item in body["data"])
@pytest.mark.asyncio
async def test_list_permissions_without_token(client):
    resp = await client.get("/api/v1/auth/permissions")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_guest_cannot_view_permissions(client, guest_token):
    """guest 缺少 auth:permission:view，应返回 20051。"""
    resp = await client.get(
        "/api/v1/auth/permissions",
        headers={"Authorization": f"Bearer {guest_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
@pytest.mark.asyncio
async def test_get_module_permissions_invalid_module_id(client, admin_token):
    """v1.1 P2-5：非法 module_id 返回 90001。"""
    resp = await client.get(
        "/api/v1/auth/permissions/modules/BadModule",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001
# ===========================================================================
# 内部注册函数：正常路径
# ===========================================================================
@pytest.mark.asyncio
async def test_register_permissions_idempotent():
    """register_permissions 幂等：首次新增，二次仅更新 name。"""
    from src.modules.auth.permission_service import register_permissions
    module_id = "test_permission_module"
    code = f"{module_id}:demo:view"
    async with AsyncSessionLocal() as session:
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
        first = await register_permissions(
            session,
            module_id,
            [
                {
                    "code": code,
                    "name": "测试查看",
                    "resource": "demo",
                    "action": "view",
                }
            ],
        )
        assert first["created"] == 1
        assert first["updated"] == 0
        assert first["total"] == 1
        second = await register_permissions(
            session,
            module_id,
            [
                {
                    "code": code,
                    "name": "测试查看2",
                    "resource": "demo",
                    "action": "view",
                }
            ],
        )
        assert second["created"] == 0
        assert second["updated"] == 1
        assert second["total"] == 1
        perm = await session.scalar(
            select(Permission).where(Permission.code == code)
        )
        assert perm is not None
        assert perm.name == "测试查看2"
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
@pytest.mark.asyncio
async def test_register_permissions_empty_list():
    """v1.1 P1-3：空列表正常返回。"""
    from src.modules.auth.permission_service import register_permissions
    async with AsyncSessionLocal() as session:
        result = await register_permissions(session, "test_empty_module", [])
        assert result["created"] == 0
        assert result["updated"] == 0
        assert result["total"] == 0
# ===========================================================================
# 内部注册函数：异常路径
# ===========================================================================
def test_validate_permission_item_invalid_code():
    """权限编码格式非法时返回 20054。"""
    from src.modules.auth.permission_service import _validate_permission_item
    with pytest.raises(PlatformException) as exc_info:
        _validate_permission_item(
            "demo",
            {
                "code": "invalid_code",
                "name": "非法",
                "resource": "demo",
                "action": "view",
            },
        )
    assert exc_info.value.code == 20054
def test_validate_permission_item_non_string_field():
    """v1.1 P0-2：非字符串字段返回 90001，而不是 TypeError。"""
    from src.modules.auth.permission_service import _validate_permission_item
    with pytest.raises(PlatformException) as exc_info:
        _validate_permission_item(
            "demo",
            {
                "code": "demo:demo:view",
                "name": 123,  # 非字符串
                "resource": "demo",
                "action": "view",
            },
        )
    assert exc_info.value.code == 90001
def test_validate_permission_item_uppercase_code():
    """v1.1 P1-2：大写权限编码返回 20054。"""
    from src.modules.auth.permission_service import _validate_permission_item
    with pytest.raises(PlatformException) as exc_info:
        _validate_permission_item(
            "demo",
            {
                "code": "demo:Demo:view",
                "name": "非法",
                "resource": "Demo",
                "action": "view",
            },
        )
    assert exc_info.value.code == 20054
def test_validate_module_id_uppercase():
    """v1.1 P1-1：大写 module_id 返回 90001。"""
    from src.modules.auth.permission_service import _validate_module_id
    with pytest.raises(PlatformException) as exc_info:
        _validate_module_id("BadModule")
    assert exc_info.value.code == 90001
@pytest.mark.asyncio
async def test_register_permissions_duplicate_code():
    """v1.1 P0-3：重复 code 返回 90001（HTTP 400）。"""
    from src.modules.auth.permission_service import register_permissions
    module_id = "test_dup_module"
    code = f"{module_id}:demo:view"
    async with AsyncSessionLocal() as session:
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
        with pytest.raises(PlatformException) as exc_info:
            await register_permissions(
                session,
                module_id,
                [
                    {
                        "code": code,
                        "name": "重复1",
                        "resource": "demo",
                        "action": "view",
                    },
                    {
                        "code": code,
                        "name": "重复2",
                        "resource": "demo",
                        "action": "view",
                    },
                ],
            )
        assert exc_info.value.code == 90001
        # 确认无残留
        perm = await session.scalar(
            select(Permission).where(Permission.code == code)
        )
        assert perm is None
@pytest.mark.asyncio
async def test_register_permissions_cross_module_conflict():
    """
    v1.2 P1-NEW-1：跨模块冲突返回 20053（HTTP 409）。
    构造方式：直接向数据库插入一条"脏数据"——code 前缀与 module_id
    不一致的记录，再调用 register_permissions 触发 20053 分支。
    说明：正常 API 路径无法到达此分支，因为 _validate_permission_item
    会强制 code 前缀与传入的 module_id 一致。此分支为防御性代码，
    仅对历史脏数据生效（P2-NEW-5）。
    """
    from src.modules.auth.permission_service import register_permissions
    module_a = "test_module_a"
    module_b = "test_module_b"
    code = f"{module_a}:demo:view"
    async with AsyncSessionLocal() as session:
        # 清理
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
        # 直接插入"脏数据"：code 前缀是 test_module_a，但 module_id 是 test_module_b
        dirty = Permission(
            code=code,
            name="脏数据权限",
            module_id=module_b,
            resource="demo",
            action="view",
        )
        session.add(dirty)
        await session.commit()
        # 模块 A 尝试注册同一 code → 命中 20053
        with pytest.raises(PlatformException) as exc_info:
            await register_permissions(
                session,
                module_a,
                [
                    {
                        "code": code,
                        "name": "A 的权限",
                        "resource": "demo",
                        "action": "view",
                    }
                ],
            )
        assert exc_info.value.code == 20053
        # 确认脏数据未被修改
        perm = await session.scalar(
            select(Permission).where(Permission.code == code)
        )
        assert perm is not None
        assert perm.module_id == module_b
        assert perm.name == "脏数据权限"
        # 清理
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
@pytest.mark.asyncio
async def test_register_permissions_partial_rollback():
    """
    v1.2 P1-NEW-2：部分新增后冲突，commit=True 时整体回滚。
    构造：
    1. 插入一条脏数据（module_id 与 code 前缀不一致）；
    2. 让目标模块注册两个权限，第一个正常，第二个命中脏数据的 code；
    3. 期望整体回滚，第一个权限也未被落库。
    """
    from src.modules.auth.permission_service import register_permissions
    module_target = "test_rollback_target"
    module_dirty = "test_rollback_dirty"
    code_ok = f"{module_target}:ok:view"
    code_dirty = f"{module_target}:dirty:view"
    async with AsyncSessionLocal() as session:
        # 清理
        await session.execute(
            delete(Permission).where(Permission.code.in_([code_ok, code_dirty]))
        )
        await session.commit()
        # 插入脏数据：code 前缀是 target，module_id 是 dirty
        session.add(
            Permission(
                code=code_dirty,
                name="脏数据",
                module_id=module_dirty,
                resource="dirty",
                action="view",
            )
        )
        await session.commit()
        # target 注册两个权限：第一个正常，第二个命中脏数据
        with pytest.raises(PlatformException) as exc_info:
            await register_permissions(
                session,
                module_target,
                [
                    {
                        "code": code_ok,
                        "name": "OK 权限",
                        "resource": "ok",
                        "action": "view",
                    },
                    {
                        "code": code_dirty,
                        "name": "冲突权限",
                        "resource": "dirty",
                        "action": "view",
                    },
                ],
            )
        assert exc_info.value.code == 20053
        # 验证第一个权限未被落库（整体回滚）
        perm_ok = await session.scalar(
            select(Permission).where(Permission.code == code_ok)
        )
        assert perm_ok is None, "P0-1 事务原子性失败：部分新增未回滚"
        # 清理
        await session.execute(
            delete(Permission).where(Permission.code.in_([code_ok, code_dirty]))
        )
        await session.commit()
@pytest.mark.asyncio
async def test_register_permissions_commit_false_rollback():
    """
    v1.1 P0-4 / v1.2 P2-NEW-4：commit=False 时异常不影响 session 可用性。
    流程：
    4. 使用 commit=False 注册一个权限；
    5. 调用方未提交，查询当前 session 可见该权限；
    6. 调用方 rollback 后，权限未落库；
    7. session 保持可用（后续查询正常）。
    """
    from src.modules.auth.permission_service import register_permissions
    module_id = "test_commit_false"
    code = f"{module_id}:demo:view"
    async with AsyncSessionLocal() as session:
        await session.execute(delete(Permission).where(Permission.code == code))
        await session.commit()
        # commit=False 正常注册
        result = await register_permissions(
            session,
            module_id,
            [
                {
                    "code": code,
                    "name": "commit=False 测试",
                    "resource": "demo",
                    "action": "view",
                }
            ],
            commit=False,
        )
        assert result["created"] == 1
        assert result["updated"] == 0
        assert result["total"] == 1
        # 未提交，当前 session 可见
        perm = await session.scalar(
            select(Permission).where(Permission.code == code)
        )
        assert perm is not None
        # 回滚
        await session.rollback()
        # 确认未落库
        perm_after = await session.scalar(
            select(Permission).where(Permission.code == code)
        )
        assert perm_after is None
        # session 保持可用：可继续查询
        any_perm = await session.scalar(select(Permission).limit(1))
        assert any_perm is not None
