"""模块管理模块核心接口测试（v1.5）。"""
import io
import json
import uuid
import zipfile
from pathlib import Path
from unittest.mock import patch

import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy import delete, select

from src.core.config import settings
from src.core.database import AsyncSessionLocal
from src.core.exceptions import PlatformException
from src.main import app
from src.modules.module_manager.models import Module, ModuleDependency


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


def _make_module_zip(
    module_id: str,
    version: str = "1.0.0",
    router_code: str = "",
    include_manifest: bool = True,
    include_router: bool = True,
) -> bytes:
    """构造一个模块 ZIP 包。"""
    manifest = {
        "id": module_id,
        "name": f"Test {module_id}",
        "version": version,
        "description": "test module",
        "dependencies": [],
        "permissions": [
            {
                "code": f"{module_id}:demo:view",
                "name": "查看 Demo",
                "resource": "demo",
                "action": "view",
            }
        ],
        "menus": [
            {
                "id": f"{module_id}:dashboard",
                "parent_id": None,
                "title": "Demo",
                "icon": "List",
                "path": "/demo",
                "component": "views/demo.vue",
                "permission": f"{module_id}:demo:view",
                "order": 10,
            }
        ],
        "entry_backend": "router:router",
        "entry_frontend": None,
        "database_tables": [],
    }
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as zf:
        if include_manifest:
            zf.writestr(f"{module_id}/manifest.json", json.dumps(manifest))
        zf.writestr(f"{module_id}/__init__.py", "")
        if include_router:
            zf.writestr(
                f"{module_id}/router.py",
                router_code
                or (
                    "from fastapi import APIRouter\n"
                    "router = APIRouter(tags=['demo'])\n"
                    "@router.get('/ping')\n"
                    "async def ping():\n"
                    "    return {'pong': True}\n"
                ),
            )
    return buf.getvalue()


async def _cleanup_module(module_id: str) -> None:
    async with AsyncSessionLocal() as session:
        await session.execute(
            delete(ModuleDependency).where(ModuleDependency.module_id == module_id)
        )
        await session.execute(
            delete(ModuleDependency).where(ModuleDependency.dependency_id == module_id)
        )
        await session.execute(delete(Module).where(Module.id == module_id))
        await session.commit()
    target = Path(settings.MODULES_DIR) / f"module_{module_id}"
    if not target.is_absolute():
        target = Path.cwd() / target
    if target.exists():
        import shutil

        shutil.rmtree(target, ignore_errors=True)
    # 清理可能的 .new / .old / .bak
    for suffix in (".new", ".old", ".bak"):
        p = target.with_name(target.name + suffix)
        if p.exists():
            import shutil

            shutil.rmtree(p, ignore_errors=True)


# ---------------------------------------------------------------------------
# 基础接口
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_list_modules_success(client, admin_token):
    resp = await client.get(
        "/api/v1/modules?page=1&page_size=10",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["code"] == 0
    assert "items" in body["data"]
    assert body["data"]["total"] >= 1


@pytest.mark.asyncio
async def test_list_modules_without_token(client):
    resp = await client.get("/api/v1/modules")
    assert resp.status_code == 401
    assert resp.json()["code"] == 10001


@pytest.mark.asyncio
async def test_list_modules_invalid_status(client, admin_token):
    resp = await client.get(
        "/api/v1/modules?status=weird",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 90001


# ---------------------------------------------------------------------------
# 核心模块保护
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_disable_core_module_forbidden(client, admin_token):
    resp = await client.post(
        "/api/v1/modules/auth/disable",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 30013


@pytest.mark.asyncio
async def test_enable_core_module_forbidden(client, admin_token):
    resp = await client.post(
        "/api/v1/modules/auth/enable",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 30013


@pytest.mark.asyncio
async def test_uninstall_core_module_forbidden(client, admin_token):
    resp = await client.delete(
        "/api/v1/modules/auth",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 30013


# ---------------------------------------------------------------------------
# 安装
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_install_via_upload_success(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        resp = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        assert resp.status_code == 200, resp.text
        body = resp.json()
        assert body["code"] == 0
        assert body["data"]["id"] == module_id
        assert body["data"]["status"] == "inactive"

        target = Path.cwd() / Path(settings.MODULES_DIR) / f"module_{module_id}"
        assert target.exists()
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_install_duplicate_module(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        resp1 = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        assert resp1.status_code == 200

        resp2 = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        assert resp2.status_code == 409
        assert resp2.json()["code"] == 30002
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_install_invalid_zip(client, admin_token):
    """V11-P1-05：精确断言 400 / 30012。"""
    resp = await client.post(
        "/api/v1/modules/upload",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"file": ("bad.zip", b"not a real zip", "application/zip")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 30012


@pytest.mark.asyncio
async def test_install_failure_no_filesystem_residue(client, admin_token):
    """V11-P1-06：manifest 校验失败时无文件残留。"""
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    bad_manifest = {
        "id": module_id,
        "name": "Bad",
        "version": "1.0.0",
        "description": "x",
        "permissions": [
            {
                "code": "other:demo:view",
                "name": "x",
                "resource": "demo",
                "action": "view",
            }
        ],
        "menus": [],
        "entry_backend": "router:router",
    }
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr(f"{module_id}/manifest.json", json.dumps(bad_manifest))
        zf.writestr(
            f"{module_id}/router.py",
            "from fastapi import APIRouter\nrouter = APIRouter()\n",
        )

    try:
        resp = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", buf.getvalue(), "application/zip")},
        )
        assert resp.status_code == 400
        assert resp.json()["code"] == 30004

        target = Path.cwd() / Path(settings.MODULES_DIR) / f"module_{module_id}"
        assert not target.exists()
        assert not target.with_name(target.name + ".new").exists()
        assert not target.with_name(target.name + ".old").exists()

        async with AsyncSessionLocal() as session:
            module = await session.get(Module, module_id)
            assert module is None
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_install_commit_failure_no_permission_residue(client, admin_token):
    """
    V13-P0-02 / V14-P0-01：mock _swap_new_to_target 失败，验证补偿清理。
    与 test_install_failure_no_filesystem_residue 不同，本测试真正走到
    FS 替换失败路径（由于延迟提交，此时 DB 尚未 commit）。
    """
    from src.modules.auth.models import Permission

    module_id = f"t_{uuid.uuid4().hex[:8]}"
    content = _make_module_zip(module_id)

    try:
        with patch(
            "src.modules.module_manager.service._swap_new_to_target",
            side_effect=PlatformException(
                code=90000, message="mock swap failure", status_code=500
            ),
        ):
            resp = await client.post(
                "/api/v1/modules/upload",
                headers={"Authorization": f"Bearer {admin_token}"},
                files={"file": (f"{module_id}.zip", content, "application/zip")},
            )
        assert resp.status_code == 500
        assert resp.json()["code"] == 90000

        # DB 无残留
        async with AsyncSessionLocal() as session:
            m = await session.get(Module, module_id)
            assert m is None
            perm = await session.scalar(
                select(Permission).where(Permission.module_id == module_id)
            )
            assert perm is None

        # FS 无残留
        target = Path.cwd() / Path(settings.MODULES_DIR) / f"module_{module_id}"
        assert not target.exists()
        assert not target.with_name(target.name + ".new").exists()
        assert not target.with_name(target.name + ".old").exists()
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_install_manifest_permission_prefix_mismatch(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    bad_manifest = {
        "id": module_id,
        "name": "Bad",
        "version": "1.0.0",
        "description": "x",
        "permissions": [
            {
                "code": "other:demo:view",
                "name": "x",
                "resource": "demo",
                "action": "view",
            }
        ],
        "menus": [],
        "entry_backend": "router:router",
    }
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr(f"{module_id}/manifest.json", json.dumps(bad_manifest))
        zf.writestr(
            f"{module_id}/router.py",
            "from fastapi import APIRouter\nrouter = APIRouter()\n",
        )
    resp = await client.post(
        "/api/v1/modules/upload",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"file": (f"{module_id}.zip", buf.getvalue(), "application/zip")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 30004


@pytest.mark.asyncio
async def test_install_missing_entry_router(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    bad_manifest = {
        "id": module_id,
        "name": "Bad",
        "version": "1.0.0",
        "description": "x",
        "permissions": [],
        "menus": [],
        "entry_backend": "not_exist:router",
    }
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr(f"{module_id}/manifest.json", json.dumps(bad_manifest))
        zf.writestr(f"{module_id}/router.py", "")
    resp = await client.post(
        "/api/v1/modules/upload",
        headers={"Authorization": f"Bearer {admin_token}"},
        files={"file": (f"{module_id}.zip", buf.getvalue(), "application/zip")},
    )
    assert resp.status_code == 400
    assert resp.json()["code"] == 30008


@pytest.mark.asyncio
async def test_install_zip_cleaned_after_install(client, admin_token):
    """V11-P0-05 / V13-P1-08：安装成功后上传 ZIP 应被清理。"""
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    def _count_zips():
        return len(list(upload_root.glob("*.zip")))

    before_count = _count_zips()

    try:
        content = _make_module_zip(module_id)
        resp = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        assert resp.status_code == 200

        after_count = _count_zips()
        assert after_count == before_count, (
            f"上传 ZIP 未清理：{before_count} → {after_count}"
        )
    finally:
        await _cleanup_module(module_id)


# ---------------------------------------------------------------------------
# 启用 / 停用
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_enable_module_loads_router(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        resp = await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        assert resp.status_code == 200

        resp = await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 200, resp.text
        assert resp.json()["data"]["status"] == "active"

        resp = await client.get(
            f"/api/v1/{module_id}/ping",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 200, resp.text
        assert resp.json() == {"pong": True}
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_enable_idempotent_no_duplicate_routes(client, admin_token):
    """V11-P1-07：启用-停用-再启用不重复注册路由。"""
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )

        def _count_module_routes():
            return len(
                [
                    r
                    for r in app.routes
                    if hasattr(r, "path") and f"/api/v1/{module_id}" in r.path
                ]
            )

        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        count_1 = _count_module_routes()
        assert count_1 >= 1

        await client.post(
            f"/api/v1/modules/{module_id}/disable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        count_2 = _count_module_routes()
        assert count_2 == count_1, f"路由重复注册：{count_1} → {count_2}"
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_disable_module(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        resp = await client.post(
            f"/api/v1/modules/{module_id}/disable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 200
        assert resp.json()["data"]["status"] == "inactive"
    finally:
        await _cleanup_module(module_id)


# ---------------------------------------------------------------------------
# 卸载
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_uninstall_nonexistent_module(client, admin_token):
    resp = await client.delete(
        "/api/v1/modules/not_exist_module",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 404
    assert resp.json()["code"] == 30003


@pytest.mark.asyncio
async def test_uninstall_active_module_forbidden(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        resp = await client.delete(
            f"/api/v1/modules/{module_id}",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 409
        assert resp.json()["code"] == 30011
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_uninstall_cleans_permissions(client, admin_token):
    from src.modules.auth.models import Permission

    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        async with AsyncSessionLocal() as session:
            perm = await session.scalar(
                select(Permission).where(Permission.module_id == module_id)
            )
            assert perm is not None

        resp = await client.delete(
            f"/api/v1/modules/{module_id}",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert resp.status_code == 200

        async with AsyncSessionLocal() as session:
            perm = await session.scalar(
                select(Permission).where(Permission.module_id == module_id)
            )
            assert perm is None
    finally:
        await _cleanup_module(module_id)


@pytest.mark.asyncio
async def test_uninstall_then_reinstall_then_enable(client, admin_token):
    """V12-P0-04：卸载后重装同名模块，启用后 API 可访问。"""
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        content = _make_module_zip(module_id)

        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        r = await client.get(
            f"/api/v1/{module_id}/ping",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200

        await client.post(
            f"/api/v1/modules/{module_id}/disable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        r = await client.delete(
            f"/api/v1/modules/{module_id}?force=true",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200

        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", content, "application/zip")},
        )
        r = await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200

        r = await client.get(
            f"/api/v1/{module_id}/ping",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200
        assert r.json() == {"pong": True}
    finally:
        await _cleanup_module(module_id)


# ---------------------------------------------------------------------------
# 升级
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_upgrade_module(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    try:
        v1 = _make_module_zip(module_id, version="1.0.0")
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", v1, "application/zip")},
        )

        v2 = _make_module_zip(module_id, version="1.1.0")
        v2_path = upload_root / f"{module_id}_110.zip"
        v2_path.write_bytes(v2)

        resp = await client.post(
            f"/api/v1/modules/{module_id}/upgrade",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={"install_type": "zip", "file_path": str(v2_path)},
        )
        assert resp.status_code == 200, resp.text
        assert resp.json()["data"]["version"] == "1.1.0"
    finally:
        await _cleanup_module(module_id)
        for f in upload_root.glob(f"{module_id}_*.zip"):
            f.unlink(missing_ok=True)


@pytest.mark.asyncio
async def test_upgrade_to_lower_version_forbidden(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    try:
        v1 = _make_module_zip(module_id, version="2.0.0")
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", v1, "application/zip")},
        )

        v2 = _make_module_zip(module_id, version="1.0.0")
        v2_path = upload_root / f"{module_id}_100.zip"
        v2_path.write_bytes(v2)

        resp = await client.post(
            f"/api/v1/modules/{module_id}/upgrade",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={"install_type": "zip", "file_path": str(v2_path)},
        )
        assert resp.status_code == 400
        assert resp.json()["code"] == 30015
    finally:
        await _cleanup_module(module_id)
        for f in upload_root.glob(f"{module_id}_*.zip"):
            f.unlink(missing_ok=True)


@pytest.mark.asyncio
async def test_upgrade_active_module_reloads_router(client, admin_token):
    """V12-P0-03：升级 active 模块后，新版本 API 立即可用（不重启）。"""
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    try:
        v1_router_code = (
            "from fastapi import APIRouter\n"
            "router = APIRouter(tags=['demo'])\n"
            "@router.get('/ping')\n"
            "async def ping():\n"
            "    return {'pong': True}\n"
        )
        v1 = _make_module_zip(module_id, version="1.0.0", router_code=v1_router_code)
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", v1, "application/zip")},
        )
        await client.post(
            f"/api/v1/modules/{module_id}/enable",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        r = await client.get(
            f"/api/v1/{module_id}/ping",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200

        v2_router_code = (
            "from fastapi import APIRouter\n"
            "router = APIRouter(tags=['demo'])\n"
            "@router.get('/ping')\n"
            "async def ping():\n"
            "    return {'pong': True}\n"
            "@router.get('/version')\n"
            "async def version():\n"
            "    return {'v': '1.1.0'}\n"
        )
        v2 = _make_module_zip(module_id, version="1.1.0", router_code=v2_router_code)
        v2_path = upload_root / f"{module_id}_110.zip"
        v2_path.write_bytes(v2)

        r = await client.post(
            f"/api/v1/modules/{module_id}/upgrade",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={"install_type": "zip", "file_path": str(v2_path)},
        )
        assert r.status_code == 200, r.text

        r = await client.get(
            f"/api/v1/{module_id}/version",
            headers={"Authorization": f"Bearer {admin_token}"},
        )
        assert r.status_code == 200, r.text
        assert r.json() == {"v": "1.1.0"}
    finally:
        await _cleanup_module(module_id)
        for f in upload_root.glob(f"{module_id}_*.zip"):
            f.unlink(missing_ok=True)


@pytest.mark.asyncio
async def test_upgrade_commit_failure_restores_role_permissions(client, admin_token):
    """
    V13-P0-01 / V14-P0-01：mock _swap_new_to_target 失败，触发延迟提交方案回滚。
    验证：DB 版本恢复旧值，角色-权限关联保留。
    """
    from src.modules.auth.models import Permission, Role, RolePermission

    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    try:
        v1 = _make_module_zip(module_id, version="1.0.0")
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", v1, "application/zip")},
        )

        async with AsyncSessionLocal() as session:
            admin_role = await session.scalar(
                select(Role).where(Role.code == "admin")
            )
            assert admin_role is not None
            perm = await session.scalar(
                select(Permission).where(
                    Permission.code == f"{module_id}:demo:view"
                )
            )
            assert perm is not None
            session.add(
                RolePermission(role_id=admin_role.id, permission_id=perm.id)
            )
            await session.commit()

        v2 = _make_module_zip(module_id, version="1.1.0")
        v2_path = upload_root / f"{module_id}_110.zip"
        v2_path.write_bytes(v2)

        with patch(
            "src.modules.module_manager.service._swap_new_to_target",
            side_effect=PlatformException(
                code=90000, message="mock swap failure", status_code=500
            ),
        ):
            resp = await client.post(
                f"/api/v1/modules/{module_id}/upgrade",
                headers={"Authorization": f"Bearer {admin_token}"},
                json={"install_type": "zip", "file_path": str(v2_path)},
            )
        assert resp.status_code == 500

        async with AsyncSessionLocal() as session:
            admin_role = await session.scalar(
                select(Role).where(Role.code == "admin")
            )
            perm = await session.scalar(
                select(Permission).where(
                    Permission.code == f"{module_id}:demo:view"
                )
            )
            assert perm is not None

            link = await session.scalar(
                select(RolePermission).where(
                    RolePermission.role_id == admin_role.id,
                    RolePermission.permission_id == perm.id,
                )
            )
            assert link is not None, "角色-权限关联应保留"

            m = await session.get(Module, module_id)
            assert m is not None
            assert m.version == "1.0.0"
    finally:
        await _cleanup_module(module_id)
        for f in upload_root.glob(f"{module_id}_*.zip"):
            f.unlink(missing_ok=True)


@pytest.mark.asyncio
async def test_upgrade_failure_cleans_new_permissions(client, admin_token):
    """
    V14-P1-03：升级 v1.0.0 → v1.1.0，v1.1.0 权限多于 v1.0.0；
    mock _swap_new_to_target 失败；
    验证补偿后仅保留 v1.0.0 权限，v1.1.0 新增权限被清理。
    """
    from src.modules.auth.models import Permission

    module_id = f"t_{uuid.uuid4().hex[:8]}"
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root.mkdir(parents=True, exist_ok=True)

    def _make_zip_with_perms(version: str, perms: list[dict]) -> bytes:
        manifest = {
            "id": module_id,
            "name": f"Test {module_id}",
            "version": version,
            "description": "test",
            "dependencies": [],
            "permissions": perms,
            "menus": [],
            "entry_backend": "router:router",
            "entry_frontend": None,
            "database_tables": [],
        }
        buf = io.BytesIO()
        with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as zf:
            zf.writestr(f"{module_id}/manifest.json", json.dumps(manifest))
            zf.writestr(f"{module_id}/__init__.py", "")
            zf.writestr(
                f"{module_id}/router.py",
                "from fastapi import APIRouter\nrouter = APIRouter()\n",
            )
        return buf.getvalue()

    try:
        v1 = _make_zip_with_perms(
            "1.0.0",
            [
                {
                    "code": f"{module_id}:demo:view",
                    "name": "查看",
                    "resource": "demo",
                    "action": "view",
                }
            ],
        )
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", v1, "application/zip")},
        )

        v2 = _make_zip_with_perms(
            "1.1.0",
            [
                {
                    "code": f"{module_id}:demo:view",
                    "name": "查看",
                    "resource": "demo",
                    "action": "view",
                },
                {
                    "code": f"{module_id}:demo:edit",
                    "name": "编辑",
                    "resource": "demo",
                    "action": "edit",
                },
            ],
        )
        v2_path = upload_root / f"{module_id}_110.zip"
        v2_path.write_bytes(v2)

        with patch(
            "src.modules.module_manager.service._swap_new_to_target",
            side_effect=PlatformException(
                code=90000, message="mock swap failure", status_code=500
            ),
        ):
            resp = await client.post(
                f"/api/v1/modules/{module_id}/upgrade",
                headers={"Authorization": f"Bearer {admin_token}"},
                json={"install_type": "zip", "file_path": str(v2_path)},
            )
        assert resp.status_code == 500

        async with AsyncSessionLocal() as session:
            codes_result = await session.execute(
                select(Permission.code).where(
                    Permission.module_id == module_id
                )
            )
            codes = {row[0] for row in codes_result.all()}
            assert f"{module_id}:demo:view" in codes
            assert f"{module_id}:demo:edit" not in codes

            m = await session.get(Module, module_id)
            assert m is not None
            assert m.version == "1.0.0"
    finally:
        await _cleanup_module(module_id)
        for f in upload_root.glob(f"{module_id}_*.zip"):
            f.unlink(missing_ok=True)


# ---------------------------------------------------------------------------
# 配置
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_get_module_config(client, admin_token):
    resp = await client.get(
        "/api/v1/modules/module_manager/config",
        headers={"Authorization": f"Bearer {admin_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["code"] == 0


@pytest.mark.asyncio
async def test_update_module_config_bool_not_integer(client, admin_token):
    module_id = f"t_{uuid.uuid4().hex[:8]}"
    try:
        manifest = {
            "id": module_id,
            "name": "Cfg",
            "version": "1.0.0",
            "description": "x",
            "permissions": [],
            "menus": [],
            "entry_backend": "router:router",
            "config_schema": {
                "type": "object",
                "properties": {
                    "max_customers": {"type": "integer"},
                },
            },
        }
        buf = io.BytesIO()
        with zipfile.ZipFile(buf, "w") as zf:
            zf.writestr(f"{module_id}/manifest.json", json.dumps(manifest))
            zf.writestr(
                f"{module_id}/router.py",
                "from fastapi import APIRouter\nrouter = APIRouter()\n",
            )
        await client.post(
            "/api/v1/modules/upload",
            headers={"Authorization": f"Bearer {admin_token}"},
            files={"file": (f"{module_id}.zip", buf.getvalue(), "application/zip")},
        )

        resp = await client.put(
            f"/api/v1/modules/{module_id}/config",
            headers={"Authorization": f"Bearer {admin_token}"},
            json={"config": {"max_customers": True}},
        )
        assert resp.status_code == 400
        assert resp.json()["code"] == 30007
    finally:
        await _cleanup_module(module_id)


# ---------------------------------------------------------------------------
# 权限校验
# ---------------------------------------------------------------------------
@pytest.mark.asyncio
async def test_guest_cannot_install(client, guest_token):
    resp = await client.post(
        "/api/v1/modules/upload",
        headers={"Authorization": f"Bearer {guest_token}"},
        files={"file": ("x.zip", b"", "application/zip")},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051


@pytest.mark.asyncio
async def test_guest_cannot_uninstall(client, guest_token):
    resp = await client.delete(
        "/api/v1/modules/auth",
        headers={"Authorization": f"Bearer {guest_token}"},
    )
    assert resp.status_code == 403
    assert resp.json()["code"] == 20051
