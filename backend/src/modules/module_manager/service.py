"""模块管理模块 - 业务逻辑（v1.5）。"""
import copy
import logging
import os
import re
import shutil
import tempfile
import uuid
import zipfile
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from fastapi import FastAPI
from packaging.version import InvalidVersion, Version
from sqlalchemy import delete, func, or_, select, text as sa_text
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from src.core.config import settings
from src.core.exceptions import PlatformException
from src.modules.auth.models import Permission, RolePermission
from src.modules.auth.permission_service import (
    register_permissions,
    unregister_permissions,
)
from src.modules.auth.service import log_auth_event
from src.modules.module_manager.constants import (
    COPY_IGNORE_PATTERNS,
    CORE_MODULE_IDS,
    CORE_MODULES,
    DEFAULT_ZIP_MAX_FILES,
    DEFAULT_ZIP_MAX_SIZE,
    DEFAULT_ZIP_MAX_TOTAL,
)
from src.modules.module_manager.loader import (
    get_module_dir,
    load_single_module,
    unmark_module_loaded,
)
from src.modules.module_manager.manifest import (
    load_manifest_from_dir,
    validate_manifest_dict,
)
from src.modules.module_manager.models import Module, ModuleDependency

logger = logging.getLogger(__name__)

# ZIP 压缩比上限
_ZIP_MAX_COMPRESS_RATIO = 100
# 表名格式
_TABLE_NAME_PATTERN = re.compile(r"^[a-z][a-z0-9_]*$")


def _utcnow_naive() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


# ---------------------------------------------------------------------------
# 序列化
# ---------------------------------------------------------------------------
def _serialize_module(module: Module) -> Dict[str, Any]:
    """将 Module ORM 序列化为统一字典。"""
    manifest = module.manifest or {}
    raw_menus = manifest.get("menus", []) or []
    menus: List[Dict[str, Any]] = []
    for item in raw_menus:
        menus.append(
            {
                "id": str(item.get("id", "")),
                "parent_id": item.get("parent_id"),
                "title": str(item.get("title", "")),
                "icon": item.get("icon"),
                "path": str(item.get("path", "")),
                "component": str(item.get("component", "")),
                "permission": item.get("permission"),
                "order": int(item.get("order", 0)),
            }
        )
    return {
        "id": module.id,
        "name": module.name,
        "version": module.version,
        "description": module.description,
        "author": module.author,
        "homepage": module.homepage,
        "status": module.status,
        "entry_backend": module.entry_backend,
        "entry_frontend": module.entry_frontend,
        "dependencies": [d.dependency_id for d in module.dependencies],
        "menus": menus,
        "config": module.config or {},
        "installed_at": module.installed_at,
        "updated_at": module.updated_at,
    }


def filter_menus_by_permission(
    menus: List[Dict[str, Any]], user_permissions: List[str]
) -> List[Dict[str, Any]]:
    """按权限递归过滤菜单。"""
    if not menus:
        return []

    by_id = {m["id"]: m for m in menus}
    children: Dict[str, List[Dict[str, Any]]] = {}
    roots: List[Dict[str, Any]] = []

    for m in menus:
        pid = m.get("parent_id")
        if pid and pid in by_id:
            children.setdefault(pid, []).append(m)
        else:
            roots.append(m)

    allowed: set[str] = set()

    def _visit(node: Dict[str, Any]) -> None:
        perm = node.get("permission")
        if perm and perm not in user_permissions:
            return
        allowed.add(node["id"])
        for child in children.get(node["id"], []):
            _visit(child)

    for root in roots:
        _visit(root)

    return [m for m in menus if m["id"] in allowed]


# ---------------------------------------------------------------------------
# 通用工具
# ---------------------------------------------------------------------------
async def _load_module_with_deps(
    db: AsyncSession, module_id: str
) -> Optional[Module]:
    return await db.scalar(
        select(Module)
        .options(selectinload(Module.dependencies))
        .where(Module.id == module_id)
    )


def _ensure_not_core_module(module_id: str, action: str) -> None:
    if module_id in CORE_MODULE_IDS:
        raise PlatformException(
            code=30013,
            message=f"核心模块不允许执行 {action} 操作",
            status_code=403,
        )


def _safe_rmtree_quiet(path: Path) -> None:
    """删除目录，失败仅记日志。"""
    if not path.exists():
        return
    try:
        shutil.rmtree(path)
    except OSError as exc:
        logger.warning("清理目录失败（忽略）：%s，原因：%s", path, exc)


# ---------------------------------------------------------------------------
# FS 三工具函数（V14-P0-01 核心）
# ---------------------------------------------------------------------------
def _prepare_new_dir(source: Path, target: Path) -> Path:
    """复制源码到 <target>.new。"""
    new_dir = target.with_name(target.name + ".new")
    if new_dir.exists():
        _safe_rmtree_quiet(new_dir)
    try:
        shutil.copytree(
            source,
            new_dir,
            ignore=shutil.ignore_patterns(*COPY_IGNORE_PATTERNS),
        )
    except OSError as exc:
        _safe_rmtree_quiet(new_dir)
        raise PlatformException(
            code=90000, message=f"模块源码复制失败：{exc}", status_code=500
        ) from exc

    if not (new_dir / "manifest.json").exists():
        _safe_rmtree_quiet(new_dir)
        raise PlatformException(
            code=30004, message="模块源码不完整（缺少 manifest.json）", status_code=400
        )
    return new_dir


def _swap_new_to_target(new_dir: Path, target: Path) -> Optional[Path]:
    """
    将 new_dir 替换为 target，保留旧版本为 target.old。

    返回 old_backup 路径（若 target 之前不存在则返回 None）。
    失败时恢复旧目录、清理 new_dir，并抛出 PlatformException。
    """
    old_backup = target.with_name(target.name + ".old")
    if old_backup.exists():
        _safe_rmtree_quiet(old_backup)

    target_existed = target.exists()

    try:
        if target_existed:
            os.replace(target, old_backup)
        os.replace(new_dir, target)
    except OSError as exc:
        if target_existed and old_backup.exists() and not target.exists():
            try:
                os.replace(old_backup, target)
            except OSError as restore_exc:
                logger.critical(
                    "恢复旧目录失败：%s → %s，原因：%s，需人工介入",
                    old_backup,
                    target,
                    restore_exc,
                )
        _safe_rmtree_quiet(new_dir)
        raise PlatformException(
            code=90000, message=f"模块目录替换失败：{exc}", status_code=500
        ) from exc

    return old_backup if target_existed else None


def _finalize_swap(old_backup: Optional[Path]) -> None:
    """成功后删除旧版本备份。"""
    if old_backup and old_backup.exists():
        _safe_rmtree_quiet(old_backup)


def _rollback_swap(target: Path, old_backup: Optional[Path]) -> None:
    """
    回滚 FS：删除 target（若存在），恢复 old_backup 到 target。
    - install：old_backup=None，删除 target，回到初始
    - upgrade：old_backup=target.old，删除 target，恢复旧版本
    """
    if target.exists():
        _safe_rmtree_quiet(target)
    if old_backup and old_backup.exists():
        try:
            os.replace(old_backup, target)
        except OSError as exc:
            logger.critical(
                "恢复旧目录失败：%s → %s，原因：%s，需人工介入",
                old_backup,
                target,
                exc,
            )


# ---------------------------------------------------------------------------
# 路径与 ZIP 校验
# ---------------------------------------------------------------------------
def _validate_source_path(source_path: str) -> Path:
    raw = Path(source_path)
    if not raw.is_absolute():
        raw = Path.cwd() / raw

    resolved = raw.resolve()
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root = upload_root.resolve()

    try:
        resolved.relative_to(upload_root)
    except ValueError:
        raise PlatformException(
            code=90001,
            message=f"source_path 必须位于 {settings.MODULE_UPLOAD_DIR} 下",
            status_code=400,
        )

    if not resolved.exists() or not resolved.is_dir():
        raise PlatformException(
            code=90001, message="source_path 必须为已存在目录", status_code=400
        )
    return resolved


def _safe_extract_zip(
    zip_path: Path,
    target_dir: Path,
    *,
    max_size: int,
    max_total: int,
    max_files: int,
) -> None:
    """安全解压 ZIP。"""
    target_dir.mkdir(parents=True, exist_ok=True)
    target_root = target_dir.resolve()

    total_size = 0
    file_count = 0

    with zipfile.ZipFile(zip_path, "r") as zf:
        for info in zf.infolist():
            file_count += 1
            if file_count > max_files:
                raise PlatformException(
                    code=30012,
                    message=f"ZIP 包文件数量超过上限 {max_files}",
                    status_code=400,
                )
            if info.file_size > max_size:
                raise PlatformException(
                    code=30012,
                    message=f"ZIP 包内单文件超过大小上限 {max_size}",
                    status_code=400,
                )

            if info.compress_size > 0:
                ratio = info.file_size / info.compress_size
                if ratio > _ZIP_MAX_COMPRESS_RATIO:
                    raise PlatformException(
                        code=30012,
                        message=f"ZIP 包内文件压缩比异常：{info.filename}",
                        status_code=400,
                    )

            total_size += info.file_size
            if total_size > max_total:
                raise PlatformException(
                    code=30012,
                    message=f"ZIP 包解压后总大小超过上限 {max_total}",
                    status_code=400,
                )

            mode = (info.external_attr >> 16) & 0o170000
            if mode == 0o120000:
                raise PlatformException(
                    code=30012,
                    message=f"ZIP 包包含符号链接：{info.filename}",
                    status_code=400,
                )

            name = info.filename.replace("\\", "/")
            member_path = (target_root / name).resolve()
            try:
                member_path.relative_to(target_root)
            except ValueError:
                raise PlatformException(
                    code=30012,
                    message=f"ZIP 包包含非法路径：{info.filename}",
                    status_code=400,
                )

        zf.extractall(target_dir)


def _resolve_module_source(
    install_type: str,
    file_path: Optional[str],
    source_path: Optional[str],
    temp_dir: Path,
) -> Path:
    if install_type == "zip":
        if not file_path:
            raise PlatformException(
                code=90001, message="file_path 不能为空", status_code=400
            )
        zip_path = Path(file_path)
        if not zip_path.is_absolute():
            zip_path = Path.cwd() / zip_path
        if not zip_path.exists() or not zip_path.is_file():
            raise PlatformException(
                code=30012, message="ZIP 包不存在", status_code=400
            )
        _safe_extract_zip(
            zip_path,
            temp_dir,
            max_size=settings.MODULE_ZIP_MAX_SIZE or DEFAULT_ZIP_MAX_SIZE,
            max_total=settings.MODULE_ZIP_MAX_TOTAL or DEFAULT_ZIP_MAX_TOTAL,
            max_files=settings.MODULE_ZIP_MAX_FILES or DEFAULT_ZIP_MAX_FILES,
        )
        manifests = list(temp_dir.rglob("manifest.json"))
        if len(manifests) != 1:
            raise PlatformException(
                code=30012,
                message="ZIP 包中未找到唯一 manifest.json",
                status_code=400,
            )
        return manifests[0].parent

    if install_type == "path":
        if not source_path:
            raise PlatformException(
                code=90001, message="source_path 不能为空", status_code=400
            )
        return _validate_source_path(source_path)

    raise PlatformException(
        code=90001, message="install_type 必须为 zip 或 path", status_code=400
    )


def _cleanup_uploaded_zip(file_path: Optional[str]) -> None:
    """清理上传的 ZIP（仅当位于 MODULE_UPLOAD_DIR 内）。"""
    if not file_path:
        return
    try:
        zip_path = Path(file_path)
        if not zip_path.is_absolute():
            zip_path = Path.cwd() / zip_path
        resolved = zip_path.resolve()

        upload_root = Path(settings.MODULE_UPLOAD_DIR)
        if not upload_root.is_absolute():
            upload_root = Path.cwd() / upload_root
        upload_root = upload_root.resolve()

        try:
            resolved.relative_to(upload_root)
        except ValueError:
            return

        if resolved.exists():
            resolved.unlink()
            logger.debug("已清理上传 ZIP：%s", resolved)
    except Exception as exc:
        logger.warning("清理上传 ZIP 失败：%s，原因：%s", file_path, exc)


async def _check_dependencies(db: AsyncSession, dependencies: List[str]) -> None:
    for dep in dependencies:
        dep_module = await db.get(Module, dep)
        if not dep_module:
            raise PlatformException(
                code=30001, message=f"依赖模块缺失：{dep}", status_code=400
            )
        if dep_module.status != "active":
            raise PlatformException(
                code=30001, message=f"依赖模块未启用：{dep}", status_code=400
            )


# ---------------------------------------------------------------------------
# 角色-权限关联快照 / 恢复（工具函数，供未来扩展使用）
# ---------------------------------------------------------------------------
async def _snapshot_role_permissions(
    db: AsyncSession, module_id: str
) -> List[Tuple[int, str]]:
    """查询该模块所有权限的角色关联 (role_id, permission_code)。"""
    stmt = (
        select(RolePermission.role_id, Permission.code)
        .join(Permission, Permission.id == RolePermission.permission_id)
        .where(Permission.module_id == module_id)
    )
    result = await db.execute(stmt)
    return [(r, c) for r, c in result.all()]


async def _restore_role_permissions(
    db: AsyncSession,
    pairs: List[Tuple[int, str]],
) -> None:
    """
    恢复角色-权限关联（幂等，批量）。

    v1.5（V14-P1-02）：
        - 批量查询 Permission（IN 查询）；
        - 批量查询已有 RolePermission；
        - db.add_all 一次写入。

    注意：v1.5 采用延迟提交方案后，正常路径**不再需要**此函数作为补偿；
    保留作为工具函数，供未来可能的扩展场景使用。
    """
    if not pairs:
        return

    codes = list({code for _, code in pairs})
    role_ids = list({rid for rid, _ in pairs})

    perm_result = await db.execute(
        select(Permission).where(Permission.code.in_(codes))
    )
    perm_by_code = {p.code: p for p in perm_result.scalars().all()}

    perm_ids = [p.id for p in perm_by_code.values()]
    if not perm_ids:
        return

    existing_result = await db.execute(
        select(RolePermission.role_id, RolePermission.permission_id).where(
            RolePermission.role_id.in_(role_ids),
            RolePermission.permission_id.in_(perm_ids),
        )
    )
    existing = set(existing_result.all())

    to_add = []
    for role_id, code in pairs:
        perm = perm_by_code.get(code)
        if not perm:
            logger.warning(
                "补偿时权限不存在，跳过关联恢复：%s (role=%s)", code, role_id
            )
            continue
        if (role_id, perm.id) in existing:
            continue
        to_add.append(RolePermission(role_id=role_id, permission_id=perm.id))

    if to_add:
        db.add_all(to_add)


# ---------------------------------------------------------------------------
# 种子数据
# ---------------------------------------------------------------------------
async def ensure_module_seed_data(db: AsyncSession) -> None:
    """幂等插入核心模块种子数据。"""
    for item in CORE_MODULES:
        module = await db.get(Module, item["id"])
        new_deps = item.get("manifest", {}).get("dependencies", []) or []

        if module:
            module.name = item["name"]
            module.version = item["version"]
            module.description = item["description"]
            module.author = item["author"]
            module.homepage = item["homepage"]
            module.entry_backend = item["entry_backend"]
            module.entry_frontend = item["entry_frontend"]
            module.manifest = item.get("manifest")

            await db.execute(
                delete(ModuleDependency).where(
                    ModuleDependency.module_id == item["id"]
                )
            )
            for dep in new_deps:
                db.add(
                    ModuleDependency(module_id=item["id"], dependency_id=dep)
                )
            continue

        module = Module(
            id=item["id"],
            name=item["name"],
            version=item["version"],
            description=item["description"],
            author=item["author"],
            homepage=item["homepage"],
            status=item["status"],
            entry_backend=item["entry_backend"],
            entry_frontend=item["entry_frontend"],
            config=item.get("config") or {},
            manifest=item.get("manifest"),
        )
        db.add(module)
        for dep in new_deps:
            db.add(ModuleDependency(module_id=item["id"], dependency_id=dep))

    await db.commit()
    logger.info("模块种子数据初始化完成")


# ---------------------------------------------------------------------------
# 列表
# ---------------------------------------------------------------------------
async def list_modules(
    db: AsyncSession,
    page: int,
    page_size: int,
    status: Optional[str] = None,
    keyword: Optional[str] = None,
    user_permissions: Optional[List[str]] = None,
) -> Dict[str, Any]:
    conditions = []
    if status:
        if status not in ("active", "inactive"):
            raise PlatformException(
                code=90001,
                message="status 必须为 active 或 inactive",
                status_code=400,
            )
        conditions.append(Module.status == status)
    if keyword:
        like = f"%{keyword}%"
        conditions.append(or_(Module.id.ilike(like), Module.name.ilike(like)))

    count_stmt = select(func.count(Module.id))
    if conditions:
        count_stmt = count_stmt.where(*conditions)
    total = (await db.execute(count_stmt)).scalar_one()

    stmt = select(Module).options(selectinload(Module.dependencies))
    if conditions:
        stmt = stmt.where(*conditions)
    stmt = (
        stmt.order_by(Module.installed_at.asc())
        .offset((page - 1) * page_size)
        .limit(page_size)
    )
    result = await db.execute(stmt)
    modules = list(result.scalars().all())

    items: List[Dict[str, Any]] = []
    for m in modules:
        item = _serialize_module(m)
        if user_permissions is not None:
            item["menus"] = filter_menus_by_permission(
                item["menus"], user_permissions
            )
        items.append(item)

    return {
        "items": items,
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }


# ---------------------------------------------------------------------------
# 安装（V14-P0-01：延迟提交方案）
# ---------------------------------------------------------------------------
async def install_module(
    db: AsyncSession,
    install_type: str,
    file_path: Optional[str],
    source_path: Optional[str],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    """
    安装模块（v1.5）：

        阶段 1：staging + manifest 校验 + 依赖检查
        阶段 2：_prepare_new_dir → <target>.new
        阶段 3：_swap_new_to_target → <target>（install 时 target 不存在）
        阶段 4：DB 修改（add Module + ModuleDependency + register_permissions）
        阶段 5：await db.commit()   ← 最后一步
        阶段 6：_finalize_swap（install 时为 None，无操作）

    失败补偿：
        - 阶段 1/2 失败：清理 new_dir，DB 无修改
        - 阶段 3 失败：_swap_new_to_target 内部恢复
        - 阶段 4/5 失败：db.rollback() + _rollback_swap 删除 target
    """
    temp_dir = Path(tempfile.mkdtemp(prefix="module_install_"))
    module_id: Optional[str] = None
    target_dir: Optional[Path] = None
    new_dir: Optional[Path] = None
    old_backup: Optional[Path] = None
    fs_swapped = False

    try:
        # ---- 阶段 1：staging 与校验 ----
        source_root = _resolve_module_source(
            install_type, file_path, source_path, temp_dir
        )
        raw_manifest = load_manifest_from_dir(source_root)
        manifest = validate_manifest_dict(raw_manifest, module_dir=source_root)
        module_id = manifest.id

        if module_id in CORE_MODULE_IDS:
            raise PlatformException(
                code=30002, message="模块 ID 与核心模块冲突", status_code=409
            )

        existing = await db.get(Module, module_id)
        if existing:
            raise PlatformException(
                code=30002, message="模块已存在", status_code=409
            )

        await _check_dependencies(db, manifest.dependencies)

        target_dir = get_module_dir(module_id)
        if target_dir.exists():
            raise PlatformException(
                code=30002,
                message=f"模块目录已存在但无 DB 记录：{target_dir}",
                status_code=409,
            )

        # ---- 阶段 2：准备 new_dir ----
        new_dir = _prepare_new_dir(source_root, target_dir)

        # ---- 阶段 3：FS 替换 ----
        old_backup = _swap_new_to_target(new_dir, target_dir)
        fs_swapped = True

        # ---- 阶段 4：DB 修改（不 commit）----
        module = Module(
            id=module_id,
            name=manifest.name,
            version=manifest.version,
            description=manifest.description,
            author=manifest.author,
            homepage=manifest.homepage,
            status="inactive",
            entry_backend=manifest.entry_backend,
            entry_frontend=manifest.entry_frontend,
            config={},
            manifest=raw_manifest,
        )
        db.add(module)
        for dep in manifest.dependencies:
            db.add(ModuleDependency(module_id=module_id, dependency_id=dep))

        await db.flush()

        if manifest.permissions:
            await register_permissions(
                db,
                module_id,
                [p.model_dump() for p in manifest.permissions],
                commit=False,
            )

        # ---- 阶段 5：commit（最后一步）----
        await db.commit()

        # ---- 阶段 6：finalize ----
        _finalize_swap(old_backup)
        fs_swapped = False  # 已 finalize，无需回滚

        module = await _load_module_with_deps(db, module_id)

        log_auth_event(
            "module_install",
            user_id=operator["id"],
            username=operator["username"],
            status="success",
            detail=f"安装模块 {module_id} v{manifest.version}",
        )
        logger.info("模块安装成功：%s v%s", module_id, manifest.version)
        return _serialize_module(module)

    except PlatformException as exc:
        await db.rollback()
        if fs_swapped and target_dir is not None:
            _rollback_swap(target_dir, old_backup)
        elif new_dir is not None:
            _safe_rmtree_quiet(new_dir)
        log_auth_event(
            "module_install",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=exc.code,
            detail=f"安装模块失败：{exc.message}",
        )
        raise

    except IntegrityError as exc:
        await db.rollback()
        if fs_swapped and target_dir is not None:
            _rollback_swap(target_dir, old_backup)
        elif new_dir is not None:
            _safe_rmtree_quiet(new_dir)
        log_auth_event(
            "module_install",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=30002,
            detail=f"安装模块失败：{exc}",
        )
        raise PlatformException(
            code=30002, message="模块已存在或数据冲突", status_code=409
        ) from exc

    except Exception as exc:
        await db.rollback()
        if fs_swapped and target_dir is not None:
            _rollback_swap(target_dir, old_backup)
        elif new_dir is not None:
            _safe_rmtree_quiet(new_dir)
        log_auth_event(
            "module_install",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=90000,
            detail=f"安装模块失败：{exc}",
        )
        raise

    finally:
        shutil.rmtree(temp_dir, ignore_errors=True)
        # V14-P1-06：确保 new_dir 清理（幂等）
        if new_dir is not None and new_dir.exists():
            _safe_rmtree_quiet(new_dir)
        if install_type == "zip":
            _cleanup_uploaded_zip(file_path)


# ---------------------------------------------------------------------------
# 升级（V14-P0-01：延迟提交方案；V14-P1-04：失败清理加载标记）
# ---------------------------------------------------------------------------
async def upgrade_module(
    db: AsyncSession,
    module_id: str,
    install_type: str,
    file_path: Optional[str],
    source_path: Optional[str],
    operator: Dict[str, Any],
    app: FastAPI,
) -> Dict[str, Any]:
    """
    升级模块（v1.5）：

        阶段 1：快照（在 try 内，V14-P1-05）
        阶段 2：staging + manifest 校验 + 版本比较 + 依赖检查
        阶段 3：_prepare_new_dir → <target>.new
        阶段 4：_swap_new_to_target → <target>（保留 <target>.old）
        阶段 5：DB 修改（不 commit）
        阶段 6：await db.commit()   ← 最后一步
        阶段 7：_finalize_swap 删除 <target>.old
        阶段 8：active 模块重载路由

    失败处理：
        - 阶段 5/6 失败：db.rollback() + _rollback_swap 恢复旧版本
        - 补偿失败：unmark_module_loaded(app, module_id)（V14-P1-04）
    """
    _ensure_not_core_module(module_id, "upgrade")

    module = await _load_module_with_deps(db, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    temp_dir = Path(tempfile.mkdtemp(prefix="module_upgrade_"))
    new_dir: Optional[Path] = None
    target_dir: Optional[Path] = None
    old_backup: Optional[Path] = None
    fs_swapped = False
    old_ver: Optional[Version] = None
    new_ver: Optional[Version] = None

    try:
        # ---- 阶段 1：快照（V14-P1-05：放入 try 内）----
        old_manifest = copy.deepcopy(module.manifest)
        old_dependencies = [d.dependency_id for d in module.dependencies]
        snapshot = {
            "name": module.name,
            "version": module.version,
            "description": module.description,
            "author": module.author,
            "homepage": module.homepage,
            "entry_backend": module.entry_backend,
            "entry_frontend": module.entry_frontend,
            "manifest": old_manifest,
            "status": module.status,
            "dependencies": old_dependencies,
        }

        # ---- 阶段 2：staging 与校验 ----
        source_root = _resolve_module_source(
            install_type, file_path, source_path, temp_dir
        )
        raw_manifest = load_manifest_from_dir(source_root)
        manifest = validate_manifest_dict(raw_manifest, module_dir=source_root)

        if manifest.id != module_id:
            raise PlatformException(
                code=30004,
                message=f"升级包模块 ID 不一致：{manifest.id} != {module_id}",
                status_code=400,
            )

        try:
            new_ver = Version(manifest.version)
            old_ver = Version(module.version)
        except InvalidVersion as exc:
            raise PlatformException(
                code=30004, message=f"版本号格式无效：{exc}", status_code=400
            ) from exc

        if new_ver <= old_ver:
            raise PlatformException(
                code=30015,
                message=f"新版本 {new_ver} 必须高于当前版本 {old_ver}",
                status_code=400,
            )

        await _check_dependencies(db, manifest.dependencies)

        target_dir = get_module_dir(module_id)

        # ---- 阶段 3：准备 new_dir ----
        new_dir = _prepare_new_dir(source_root, target_dir)

        # ---- 阶段 4：FS 替换（保留 old_backup）----
        old_backup = _swap_new_to_target(new_dir, target_dir)
        fs_swapped = True

        # ---- 阶段 5：DB 修改（不 commit）----
        module.name = manifest.name
        module.version = manifest.version
        module.description = manifest.description
        module.author = manifest.author
        module.homepage = manifest.homepage
        module.entry_backend = manifest.entry_backend
        module.entry_frontend = manifest.entry_frontend
        module.manifest = raw_manifest

        await db.execute(
            delete(ModuleDependency).where(ModuleDependency.module_id == module_id)
        )
        for dep in manifest.dependencies:
            db.add(ModuleDependency(module_id=module_id, dependency_id=dep))

        if manifest.permissions:
            await register_permissions(
                db,
                module_id,
                [p.model_dump() for p in manifest.permissions],
                commit=False,
            )

        # ---- 阶段 6：commit（最后一步）----
        await db.commit()

        # ---- 阶段 7：finalize ----
        _finalize_swap(old_backup)
        fs_swapped = False  # 已 finalize，无需回滚

        module = await _load_module_with_deps(db, module_id)

        # ---- 阶段 8：active 模块重载路由 ----
        if module.status == "active":
            try:
                unmark_module_loaded(app, module_id)
                await load_single_module(app, db, module_id)
            except Exception as exc:
                logger.error("升级后重载路由失败：%s", exc)

        log_auth_event(
            "module_upgrade",
            user_id=operator["id"],
            username=operator["username"],
            status="success",
            detail=f"升级模块 {module_id}：{old_ver} → {new_ver}",
        )
        logger.info("模块升级成功：%s → %s", module_id, new_ver)
        return _serialize_module(module)

    except PlatformException as exc:
        await db.rollback()
        if fs_swapped and target_dir is not None:
            _rollback_swap(target_dir, old_backup)
            # V14-P1-04：清理加载标记
            unmark_module_loaded(app, module_id)
        elif new_dir is not None:
            _safe_rmtree_quiet(new_dir)
        log_auth_event(
            "module_upgrade",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=exc.code,
            detail=f"升级模块失败：{exc.message}",
        )
        raise

    except Exception as exc:
        await db.rollback()
        if fs_swapped and target_dir is not None:
            _rollback_swap(target_dir, old_backup)
            # V14-P1-04：清理加载标记
            unmark_module_loaded(app, module_id)
        elif new_dir is not None:
            _safe_rmtree_quiet(new_dir)
        log_auth_event(
            "module_upgrade",
            user_id=operator["id"],
            username=operator["username"],
            status="fail",
            error_code=90000,
            detail=f"升级模块失败：{exc}",
        )
        raise

    finally:
        shutil.rmtree(temp_dir, ignore_errors=True)
        # V14-P1-06：确保 new_dir 清理（幂等）
        if new_dir is not None and new_dir.exists():
            _safe_rmtree_quiet(new_dir)
        # 升级时不清理上传 ZIP（用户可能希望保留升级包）


# ---------------------------------------------------------------------------
# 卸载
# ---------------------------------------------------------------------------
async def uninstall_module(
    db: AsyncSession,
    module_id: str,
    force: bool,
    drop_tables: bool,
    operator: Dict[str, Any],
    app: FastAPI,
) -> None:
    """卸载模块（v1.5）。"""
    _ensure_not_core_module(module_id, "uninstall")

    module = await _load_module_with_deps(db, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    if module.status == "active":
        raise PlatformException(
            code=30011, message="模块正在运行，请先停用", status_code=409
        )

    if not force:
        dep_exists = await db.scalar(
            select(ModuleDependency.id)
            .join(Module, Module.id == ModuleDependency.module_id)
            .where(
                ModuleDependency.dependency_id == module_id,
                Module.status == "active",
            )
            .limit(1)
        )
        if dep_exists:
            raise PlatformException(
                code=30006, message="模块被 active 模块依赖", status_code=400
            )

    manifest = module.manifest or {}

    deleted_perms = await unregister_permissions(db, module_id, commit=False)

    dropped_tables: List[str] = []
    if drop_tables and manifest.get("database_tables"):
        for table in manifest["database_tables"]:
            if not isinstance(table, str):
                continue
            if not _TABLE_NAME_PATTERN.match(table):
                logger.warning("跳过非法表名：%s", table)
                continue
            if not table.startswith(f"{module_id}_"):
                logger.warning("跳过不合规表名：%s", table)
                continue
            try:
                await db.execute(sa_text(f'DROP TABLE IF EXISTS "{table}"'))
                dropped_tables.append(table)
            except Exception as exc:
                logger.warning("删除表失败：%s，原因：%s", table, exc)

    await db.delete(module)
    await db.commit()

    unmark_module_loaded(app, module_id)

    target_dir = get_module_dir(module_id)
    _safe_rmtree_quiet(target_dir)

    log_auth_event(
        "module_uninstall",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=(
            f"卸载模块 {module_id}（权限清理 {deleted_perms}，"
            f"数据表清理 {len(dropped_tables)}）"
        ),
    )
    logger.info("模块卸载成功：%s", module_id)


# ---------------------------------------------------------------------------
# 启用
# ---------------------------------------------------------------------------
async def enable_module(
    db: AsyncSession,
    module_id: str,
    operator: Dict[str, Any],
    app: FastAPI,
) -> Dict[str, Any]:
    """启用模块（v1.5）。"""
    _ensure_not_core_module(module_id, "enable")

    module = await _load_module_with_deps(db, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    if module.status == "active":
        return _serialize_module(module)

    manifest = module.manifest or {}
    await _check_dependencies(db, manifest.get("dependencies", []))

    from src.modules.module_manager.manifest import _resolve_entry_file

    module_dir = get_module_dir(module_id)
    entry_attr = module.entry_backend.split(":", 1)[0]
    if _resolve_entry_file(module_dir, entry_attr) is None:
        raise PlatformException(
            code=30008,
            message=f"模块入口文件不存在：{entry_attr}",
            status_code=400,
        )

    module.status = "active"
    try:
        await db.commit()
    except Exception as exc:
        await db.rollback()
        raise PlatformException(
            code=90000,
            message=f"模块状态提交失败：{exc}",
            status_code=500,
        ) from exc

    module = await _load_module_with_deps(db, module_id)

    try:
        await load_single_module(app, db, module_id)
    except Exception as exc:
        logger.error("模块 %s 热加载失败：%s", module_id, exc)
        module.status = "inactive"
        await db.commit()
        unmark_module_loaded(app, module_id)
        raise PlatformException(
            code=30009,
            message=f"模块热加载失败：{exc}",
            status_code=500,
        ) from exc

    log_auth_event(
        "module_enable",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"启用模块 {module_id}",
    )
    return _serialize_module(module)


# ---------------------------------------------------------------------------
# 停用
# ---------------------------------------------------------------------------
async def disable_module(
    db: AsyncSession,
    module_id: str,
    operator: Dict[str, Any],
    app: FastAPI,
) -> Dict[str, Any]:
    """停用模块（v1.5）。"""
    _ensure_not_core_module(module_id, "disable")

    module = await _load_module_with_deps(db, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    if module.status == "inactive":
        return _serialize_module(module)

    dep_exists = await db.scalar(
        select(ModuleDependency.id)
        .join(Module, Module.id == ModuleDependency.module_id)
        .where(
            ModuleDependency.dependency_id == module_id,
            Module.status == "active",
        )
        .limit(1)
    )
    if dep_exists:
        raise PlatformException(
            code=30006,
            message="模块被其他 active 模块依赖，无法停用",
            status_code=400,
        )

    module.status = "inactive"
    await db.commit()
    module = await _load_module_with_deps(db, module_id)

    unmark_module_loaded(app, module_id)

    log_auth_event(
        "module_disable",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"停用模块 {module_id}",
    )
    return _serialize_module(module)


# ---------------------------------------------------------------------------
# 配置
# ---------------------------------------------------------------------------
async def get_module_config(db: AsyncSession, module_id: str) -> Dict[str, Any]:
    module = await db.get(Module, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)
    return module.config or {}


def _check_type(value: Any, expected: str) -> bool:
    if expected == "boolean":
        return isinstance(value, bool)
    if expected == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if expected == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    if expected == "string":
        return isinstance(value, str)
    if expected == "array":
        return isinstance(value, list)
    if expected == "object":
        return isinstance(value, dict)
    return True


def _validate_config(schema: Dict[str, Any], config: Dict[str, Any]) -> None:
    if not schema or not isinstance(schema, dict):
        return

    required = schema.get("required", [])
    if isinstance(required, list):
        for field in required:
            if field not in config:
                raise PlatformException(
                    code=30007,
                    message=f"配置缺少必填字段：{field}",
                    status_code=400,
                )

    properties = schema.get("properties", {})
    if not isinstance(properties, dict):
        return

    for key, value in config.items():
        prop = properties.get(key)
        if not isinstance(prop, dict):
            continue
        expected_type = prop.get("type")
        if not expected_type:
            continue
        types = (
            expected_type if isinstance(expected_type, list) else [expected_type]
        )
        if not any(_check_type(value, t) for t in types):
            raise PlatformException(
                code=30007,
                message=f"配置字段 {key} 类型无效，期望 {types}",
                status_code=400,
            )


async def update_module_config(
    db: AsyncSession,
    module_id: str,
    config: Dict[str, Any],
    operator: Dict[str, Any],
) -> Dict[str, Any]:
    module = await db.get(Module, module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    manifest = module.manifest or {}
    _validate_config(manifest.get("config_schema", {}), config)

    module.config = config
    await db.commit()
    module = await db.get(Module, module_id)

    log_auth_event(
        "module_config_update",
        user_id=operator["id"],
        username=operator["username"],
        status="success",
        detail=f"更新模块 {module_id} 配置",
    )
    return module.config or {}


async def save_uploaded_zip(file_name: str, content: bytes) -> Path:
    """保存上传的 ZIP 到 MODULE_UPLOAD_DIR。"""
    upload_root = Path(settings.MODULE_UPLOAD_DIR)
    if not upload_root.is_absolute():
        upload_root = Path.cwd() / upload_root
    upload_root = upload_root.resolve()
    upload_root.mkdir(parents=True, exist_ok=True)

    if not file_name.lower().endswith(".zip"):
        raise PlatformException(
            code=30012, message="仅支持 .zip 文件", status_code=400
        )

    if len(content) > (settings.MODULE_ZIP_MAX_SIZE or DEFAULT_ZIP_MAX_SIZE):
        raise PlatformException(
            code=30012,
            message=f"ZIP 文件大小超过上限 {settings.MODULE_ZIP_MAX_SIZE}",
            status_code=400,
        )

    target = upload_root / f"{uuid.uuid4().hex}.zip"
    target.write_bytes(content)
    return target
