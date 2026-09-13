"""License 管理模块核心接口测试（v1.2）。
v1.2 变更（R-5）：
- _fake_post 函数补充 **kwargs，兼容 timeout 等额外参数。
v1.1 变更：
- P1-5：test_activate_without_service 显式 patch 环境变量。
- P1-6：测试数量统一为 31 个。
- 新增 P0-2/P0-3/P0-5/P1-10 回归测试。
运行前请确保：
1. tests/conftest.py 已配置测试数据库（默认 test_app.db）
2. 环境变量 LICENSE_SECRET_KEY 已设置（默认即可）
"""
import json
import uuid
from datetime import datetime, timedelta, timezone
from unittest.mock import patch
import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete, select
from src.core.database import AsyncSessionLocal
from src.main import app
from src.modules.license.constants import CORE_MODULES_BYPASS
from src.modules.license.models import License
from src.modules.license.validator import (
    compute_signature,
    get_machine_code,
)
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
@pytest.fixture
async def clean_licenses():
    """测试前后清理所有 License 记录。"""
    async with AsyncSessionLocal() as session:
        await session.execute(delete(License))
        await session.commit()
    yield
    async with AsyncSessionLocal() as session:
        await session.execute(delete(License))
        await session.commit()
# ---------------------------------------------------------------------------
# 辅助函数：构造 License 文件内容
# ---------------------------------------------------------------------------
def _make_license_content(
    *,
    license_key: str = None,
    license_type: str = "enterprise",
    max_users=100,
    authorized_modules=None,
    machine_code=None,
    issued_at: datetime = None,
    expires_at: datetime = None,
    secret: str = None,
    tamper_signature: bool = False,
) -> bytes:
    """构造一个合法的 License 文件内容（已签名）。"""
    from src.core.config import settings
    if secret is None:
        secret = settings.LICENSE_SECRET_KEY
    if license_key is None:
        license_key = f"LIC-TEST-{uuid.uuid4().hex[:8].upper()}"
    if authorized_modules is None:
        authorized_modules = ["auth", "audit_log", "customer_relation"]
    now = datetime.now(timezone.utc).replace(tzinfo=None)
    if issued_at is None:
        issued_at = now
    if expires_at is None:
        expires_at = now + timedelta(days=365)
    payload = {
        "license_key": license_key,
        "license_type": license_type,
        "max_users": max_users,
        "authorized_modules": authorized_modules,
        "machine_code": machine_code,
        "issued_at": issued_at.isoformat(),
        "expires_at": expires_at.isoformat(),
    }
    signature = compute_signature(payload, secret)
    if tamper_signature:
        signature = "0" * len(signature)
    return json.dumps(
        {"payload": payload, "signature": signature}, ensure_ascii=False
    ).encode("utf-8")
# ===========================================================================
# 1. 状态查询
# ===========================================================================
@pytest.mark.asyncio
async def test_status_without_license(client, admin_token, clean_licenses):
    """未导入 License → 50009。"""
    resp = await client.get(
        "/api/v1/license/status",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 50009
@pytest.mark.asyncio
async def test_status_with_valid_license(client, admin_token, clean_licenses):
    """导入合法 License 后状态查询正常。"""
    content = _make_license_content()
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("test.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 200, resp.text
    assert resp.json()["code"] == 0
    resp = await client.get(
        "/api/v1/license/status",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["is_valid"] is True
    assert data["license_type"] == "enterprise"
    assert data["is_expired"] is False
    assert data["is_expiring_soon"] is False
    assert data["max_users"] == 100
    assert data["current_users"] >= 2
    assert "auth" in data["authorized_modules"]
    assert data["days_remaining"] > 300
    # v1.1 P1-7：machine_code 字段已移除
    assert "machine_code" not in data
    assert "current_machine_code" not in data
@pytest.mark.asyncio
async def test_status_expiring_soon(client, admin_token, clean_licenses):
    """剩余 ≤7 天 → is_expiring_soon=True。"""
    expires_at = datetime.now(timezone.utc).replace(tzinfo=None) + timedelta(days=3)
    content = _make_license_content(expires_at=expires_at)
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("test.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 200, resp.text
    resp = await client.get(
        "/api/v1/license/status",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    data = resp.json()["data"]
    assert data["is_valid"] is True
    assert data["is_expiring_soon"] is True
    assert 0 <= data["days_remaining"] <= 7
# ===========================================================================
# 2. 导入主流程
# ===========================================================================
@pytest.mark.asyncio
async def test_import_valid_license(client, admin_token, clean_licenses):
    """合法 License 导入成功。"""
    content = _make_license_content()
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("test.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    assert body["data"]["license_type"] == "enterprise"
    assert body["data"]["max_users"] == 100
    assert body["data"]["authorized_modules"] == [
        "auth",
        "audit_log",
        "customer_relation",
    ]
@pytest.mark.asyncio
async def test_import_invalid_json(client, admin_token, clean_licenses):
    """非法 JSON → 50001。"""
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("bad.lic", b"not a json", "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
@pytest.mark.asyncio
async def test_import_invalid_extension(client, admin_token, clean_licenses):
    """非法扩展名 → 50001。"""
    content = _make_license_content()
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("test.txt", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
@pytest.mark.asyncio
async def test_import_signature_failure(client, admin_token, clean_licenses):
    """签名失败 → 50003。"""
    content = _make_license_content(tamper_signature=True)
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("bad.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50003
@pytest.mark.asyncio
async def test_import_expired_license(client, admin_token, clean_licenses):
    """已过期 → 50002。"""
    expires_at = datetime.now(timezone.utc).replace(tzinfo=None) - timedelta(days=1)
    content = _make_license_content(expires_at=expires_at)
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("expired.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50002
@pytest.mark.asyncio
async def test_import_duplicate_license(client, admin_token, clean_licenses):
    """重复导入（相同 license_key 且 active）→ 50004。"""
    license_key = f"LIC-DUP-{uuid.uuid4().hex[:8].upper()}"
    content = _make_license_content(license_key=license_key)
    resp1 = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    assert resp1.status_code == 200
    resp2 = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("b.lic", content, "application/octet-stream")},
    )
    assert resp2.status_code == 409
    assert resp2.json()["code"] == 50004
@pytest.mark.asyncio
async def test_import_machine_code_mismatch(client, admin_token, clean_licenses):
    """机器码不匹配 → 50006。"""
    content = _make_license_content(machine_code="OTHER-MACHINE-CODE")
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("m.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50006
@pytest.mark.asyncio
async def test_import_invalid_license_type(client, admin_token, clean_licenses):
    """License 类型不支持 → 50010。"""
    content = _make_license_content(license_type="unknown")
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("u.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50010
@pytest.mark.asyncio
async def test_import_without_file_and_code(client, admin_token, clean_licenses):
    """既无文件也无激活码 → 50001。"""
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
# ===========================================================================
# 3. v1.1 新增：P0-5 数据校验回归测试
# ===========================================================================
@pytest.mark.asyncio
async def test_import_max_users_bool_rejected(client, admin_token, clean_licenses):
    """P0-5：max_users=True 被拒绝（bool 不是正整数）。"""
    content = _make_license_content(max_users=True)
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("b.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
@pytest.mark.asyncio
async def test_import_authorized_modules_non_string_rejected(
    client, admin_token, clean_licenses
):
    """P0-5：authorized_modules 含非字符串元素被拒绝。"""
    content = _make_license_content(authorized_modules=["auth", 123])
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("b.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
@pytest.mark.asyncio
async def test_import_license_key_too_long_rejected(
    client, admin_token, clean_licenses
):
    """P0-5：license_key 超过 255 位被拒绝。"""
    content = _make_license_content(license_key="X" * 300)
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("b.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
@pytest.mark.asyncio
async def test_import_issued_after_expires_rejected(
    client, admin_token, clean_licenses
):
    """P0-5：issued_at >= expires_at 被拒绝。"""
    now = datetime.now(timezone.utc).replace(tzinfo=None)
    content = _make_license_content(
        issued_at=now + timedelta(days=10),
        expires_at=now + timedelta(days=5),
    )
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("b.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50001
# ===========================================================================
# 4. v1.1 新增：P1-10 历史 inactive License 重新激活
# ===========================================================================
@pytest.mark.asyncio
async def test_import_inactive_license_reactivation(
    client, admin_token, clean_licenses
):
    """P1-10：历史 is_active=0 的 License 允许重新导入并激活。"""
    license_key = f"LIC-REACT-{uuid.uuid4().hex[:8].upper()}"
    content = _make_license_content(license_key=license_key)
    resp1 = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    assert resp1.status_code == 200
    async with AsyncSessionLocal() as session:
        lic = await session.scalar(
            select(License).where(License.license_key == license_key)
        )
        assert lic is not None
        lic.is_active = 0
        await session.commit()
    resp2 = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    assert resp2.status_code == 200, resp2.text
    async with AsyncSessionLocal() as session:
        lic = await session.scalar(
            select(License).where(License.license_key == license_key)
        )
        assert lic is not None
        assert lic.is_active == 1
# ===========================================================================
# 5. 在线激活
# ===========================================================================
@pytest.mark.asyncio
async def test_activate_without_service(client, admin_token, clean_licenses):
    """P1-5：显式 patch LICENSE_ACTIVATION_URL=""，未配置激活服务 → 50005。"""
    with patch("src.core.config.settings.LICENSE_ACTIVATION_URL", ""):
        resp = await client.post(
            "/api/v1/license/activate",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={
                "activation_code": "ABCD-1234-EFGH-5678",
                "machine_code": get_machine_code(),
            },
        )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50005
@pytest.mark.asyncio
async def test_activate_with_mocked_service(client, admin_token, clean_licenses):
    """模拟外部激活服务返回合法 License。"""
    content = _make_license_content()
    payload_data = json.loads(content.decode("utf-8"))
    # v1.2（R-5）：补充 **kwargs，兼容 timeout 等额外参数
    async def _fake_post(self, url, json=None, **kwargs):
        class _Resp:
            status_code = 200
            def json(self_inner):
                return {"code": 0, "data": payload_data}
        return _Resp()
    with patch("httpx.AsyncClient.post", _fake_post), patch(
        "src.core.config.settings.LICENSE_ACTIVATION_URL",
        "https://license.example.com/api",
    ):
        resp = await client.post(
            "/api/v1/license/activate",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={
                "activation_code": "ABCD-1234-EFGH-5678",
                "machine_code": get_machine_code(),
            },
        )
    assert resp.status_code == 200, resp.text
    assert resp.json()["code"] == 0
@pytest.mark.asyncio
async def test_activate_machine_code_mismatch(
    client, admin_token, clean_licenses
):
    """P0-3：请求 machine_code 与当前不一致 → 50006。"""
    with patch("src.core.config.settings.LICENSE_ACTIVATION_URL", ""):
        resp = await client.post(
            "/api/v1/license/activate",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={
                "activation_code": "ABCD-1234-EFGH-5678",
                "machine_code": "WRONG-MACHINE-CODE",
            },
        )
    assert resp.status_code == 400
    assert resp.json()["code"] == 50006
# ===========================================================================
# 6. 模块授权
# ===========================================================================
@pytest.mark.asyncio
async def test_module_authorization(client, admin_token, clean_licenses):
    """导入 License 后查询模块授权状态。"""
    content = _make_license_content(
        authorized_modules=["auth", "audit_log", "customer_relation"]
    )
    await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    resp = await client.get(
        "/api/v1/license/modules",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    items = resp.json()["data"]
    core_items = [it for it in items if it["module_id"] in CORE_MODULES_BYPASS]
    for it in core_items:
        assert it["is_authorized"] is True
    cr = next(
        (it for it in items if it["module_id"] == "customer_relation"), None
    )
    if cr is not None:
        assert cr["is_authorized"] is True
# ===========================================================================
# 7. 权限校验
# ===========================================================================
@pytest.mark.asyncio
async def test_status_without_token(client, clean_licenses):
    """未认证 → 10001。"""
    resp = await client.get("/api/v1/license/status")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001
@pytest.mark.asyncio
async def test_status_guest_forbidden(client, guest_token, clean_licenses):
    """guest 无 license:license:view → 20051。"""
    resp = await client.get(
        "/api/v1/license/status",
        headers={"Authorization": f"Bearer {guest_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
@pytest.mark.asyncio
async def test_import_guest_forbidden(client, guest_token, clean_licenses):
    """guest 无 license:license:create → 20051。"""
    content = _make_license_content()
    resp = await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {guest_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
# ===========================================================================
# 8. 用户配额
# ===========================================================================
@pytest.mark.asyncio
async def test_check_user_quota_no_license(clean_licenses):
    """无 License → 50009。"""
    from src.core.exceptions import PlatformException
    from src.modules.license.service import check_user_quota
    async with AsyncSessionLocal() as session:
        with pytest.raises(PlatformException) as exc_info:
            await check_user_quota(session)
        assert exc_info.value.code == 50009
@pytest.mark.asyncio
async def test_check_user_quota_exceeded(client, admin_token, clean_licenses):
    """max_users=1 时超限 → 50008。"""
    content = _make_license_content(max_users=1)
    await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    from src.core.exceptions import PlatformException
    from src.modules.license.service import check_user_quota
    async with AsyncSessionLocal() as session:
        with pytest.raises(PlatformException) as exc_info:
            await check_user_quota(session)
        assert exc_info.value.code == 50008
@pytest.mark.asyncio
async def test_check_user_quota_ok(client, admin_token, clean_licenses):
    """max_users=100 时不超限。"""
    content = _make_license_content(max_users=100)
    await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    from src.modules.license.service import check_user_quota
    async with AsyncSessionLocal() as session:
        await check_user_quota(session)
# ===========================================================================
# 9. 模块授权依赖
# ===========================================================================
@pytest.mark.asyncio
async def test_check_module_authorized_core_bypass(clean_licenses):
    """核心模块无需 License 直接放行。"""
    from src.modules.license.service import check_module_authorized
    async with AsyncSessionLocal() as session:
        for core in CORE_MODULES_BYPASS:
            await check_module_authorized(session, core)
@pytest.mark.asyncio
async def test_check_module_authorized_not_found(clean_licenses):
    """业务模块无 License → 50009。"""
    from src.core.exceptions import PlatformException
    from src.modules.license.service import check_module_authorized
    async with AsyncSessionLocal() as session:
        with pytest.raises(PlatformException) as exc_info:
            await check_module_authorized(session, "customer_relation")
        assert exc_info.value.code == 50009
@pytest.mark.asyncio
async def test_check_module_authorized_denied(
    client, admin_token, clean_licenses
):
    """模块未授权 → 50007。"""
    content = _make_license_content(authorized_modules=["auth"])
    await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    from src.core.exceptions import PlatformException
    from src.modules.license.service import check_module_authorized
    async with AsyncSessionLocal() as session:
        with pytest.raises(PlatformException) as exc_info:
            await check_module_authorized(session, "customer_relation")
        assert exc_info.value.code == 50007
@pytest.mark.asyncio
async def test_check_module_authorized_ok(client, admin_token, clean_licenses):
    """模块已授权时放行。"""
    content = _make_license_content(
        authorized_modules=["auth", "customer_relation"]
    )
    await client.post(
        "/api/v1/license/import",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"license_file": ("a.lic", content, "application/octet-stream")},
    )
    from src.modules.license.service import check_module_authorized
    async with AsyncSessionLocal() as session:
        await check_module_authorized(session, "customer_relation")
# ===========================================================================
# 10. v1.2 新增：R-3 并发边界（部分唯一索引）
# ===========================================================================
@pytest.mark.asyncio
async def test_partial_unique_index_prevents_multiple_active(
    client, admin_token, clean_licenses
):
    """
    v1.2（R-3）：验证部分唯一索引 uq_license_license_active。
    直接绕过 API，尝试插入两条 is_active=1 的记录，
    应触发 IntegrityError（部分唯一索引冲突）。
    说明：此测试直接操作数据库，验证数据库层约束。
    """
    from sqlalchemy.exc import IntegrityError
    async with AsyncSessionLocal() as session:
        # 第一条 active
        lic1 = License(
            license_key=f"LIC-IDX-A-{uuid.uuid4().hex[:8]}",
            license_type="enterprise",
            max_users=100,
            authorized_modules=["auth"],
            issued_at=datetime.now(timezone.utc).replace(tzinfo=None),
            expires_at=datetime.now(timezone.utc).replace(tzinfo=None)
            + timedelta(days=365),
            is_active=1,
        )
        session.add(lic1)
        await session.commit()
        # 第二条 active，应因部分唯一索引冲突失败
        lic2 = License(
            license_key=f"LIC-IDX-B-{uuid.uuid4().hex[:8]}",
            license_type="enterprise",
            max_users=100,
            authorized_modules=["auth"],
            issued_at=datetime.now(timezone.utc).replace(tzinfo=None),
            expires_at=datetime.now(timezone.utc).replace(tzinfo=None)
            + timedelta(days=365),
            is_active=1,
        )
        session.add(lic2)
        with pytest.raises(IntegrityError):
            await session.commit()
        await session.rollback()
