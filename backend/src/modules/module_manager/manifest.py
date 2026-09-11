"""模块清单 manifest.json 解析与校验（v1.3）。"""
import json
import re
from pathlib import Path
from typing import Any, Dict, List

from pydantic import BaseModel, Field, ValidationError

from src.core.exceptions import PlatformException

_MODULE_ID_PATTERN = re.compile(r"^[a-z][a-z0-9_]*$")
_PERMISSION_CODE_PATTERN = re.compile(r"^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$")
_MENU_ID_PATTERN = re.compile(r"^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$")
_SEMVER_PATTERN = re.compile(r"^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$")
_TABLE_NAME_PATTERN = re.compile(r"^[a-z][a-z0-9_]*$")
_MODULE_ATTR_PATTERN = re.compile(r"^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$")


class ManifestPermission(BaseModel):
    code: str
    name: str
    resource: str
    action: str


class ManifestMenu(BaseModel):
    id: str
    parent_id: str | None = None
    title: str
    icon: str | None = None
    path: str
    component: str
    permission: str | None = None
    order: int = 0


class Manifest(BaseModel):
    id: str
    name: str
    version: str
    description: str
    author: str | None = None
    homepage: str | None = None
    dependencies: List[str] = Field(default_factory=list)
    permissions: List[ManifestPermission] = Field(default_factory=list)
    menus: List[ManifestMenu] = Field(default_factory=list)
    config_schema: Dict[str, Any] = Field(default_factory=dict)
    entry_backend: str = "router:router"
    entry_frontend: str | None = None
    database_tables: List[str] = Field(default_factory=list)


def _resolve_entry_file(module_dir: Path, module_attr: str) -> Path | None:
    """
    解析入口文件路径（V12-P1-04：显式拒绝 .py 后缀）。
    - 支持 `router`、`api.router`、`api.router_v2`；
    - 优先 `<module_dir>/<parts>.py`，其次 `<module_dir>/<parts>/__init__.py`。
    """
    if module_attr.endswith(".py"):
        return None  # 显式拒绝
    if not _MODULE_ATTR_PATTERN.match(module_attr):
        return None

    parts = module_attr.split(".")
    candidate_py = module_dir.joinpath(*parts).with_suffix(".py")
    if candidate_py.exists():
        return candidate_py
    candidate_pkg = module_dir.joinpath(*parts) / "__init__.py"
    if candidate_pkg.exists():
        return candidate_pkg
    return None


def _validate_manifest(manifest: Manifest, module_dir: Path | None = None) -> None:
    """对 Manifest 进行业务级校验。"""
    if not _MODULE_ID_PATTERN.match(manifest.id):
        raise PlatformException(
            code=30010,
            message="模块 ID 格式无效（小写字母开头，仅允许小写字母、数字、下划线）",
            status_code=400,
        )

    if not _SEMVER_PATTERN.match(manifest.version):
        raise PlatformException(
            code=30004,
            message=f"模块版本号格式无效（应为 x.y.z）：{manifest.version}",
            status_code=400,
        )

    for dep in manifest.dependencies:
        if not _MODULE_ID_PATTERN.match(dep):
            raise PlatformException(
                code=30004, message=f"依赖模块 ID 格式无效：{dep}", status_code=400
            )

    for perm in manifest.permissions:
        if not _PERMISSION_CODE_PATTERN.match(perm.code):
            raise PlatformException(
                code=30004,
                message=f"权限编码格式无效：{perm.code}",
                status_code=400,
            )
        prefix = perm.code.split(":")[0]
        if prefix != manifest.id:
            raise PlatformException(
                code=30004,
                message=f"权限编码前缀必须与模块 ID 一致：{perm.code}",
                status_code=400,
            )

    for menu in manifest.menus:
        if not _MENU_ID_PATTERN.match(menu.id):
            raise PlatformException(
                code=30004,
                message=f"菜单 ID 格式无效（应为 {{module_id}}:{{menu_id}}）：{menu.id}",
                status_code=400,
            )
        if not menu.id.startswith(f"{manifest.id}:"):
            raise PlatformException(
                code=30004,
                message=f"菜单 ID 前缀必须为 {manifest.id}：{menu.id}",
                status_code=400,
            )

    for table in manifest.database_tables:
        if not _TABLE_NAME_PATTERN.match(table):
            raise PlatformException(
                code=30004, message=f"数据库表名格式无效：{table}", status_code=400
            )
        if not table.startswith(f"{manifest.id}_"):
            raise PlatformException(
                code=30004,
                message=f"数据库表名必须以 {manifest.id}_ 开头：{table}",
                status_code=400,
            )

    if ":" not in manifest.entry_backend:
        raise PlatformException(
            code=30008,
            message="entry_backend 格式无效，应为 {module_attr}:{router_attr}",
            status_code=400,
        )

    if module_dir is not None:
        module_attr = manifest.entry_backend.split(":", 1)[0]
        entry_file = _resolve_entry_file(module_dir, module_attr)
        if entry_file is None:
            raise PlatformException(
                code=30008,
                message=f"模块入口文件不存在：{module_attr}.py 或 {module_attr}/__init__.py",
                status_code=400,
            )


def load_manifest_from_dir(module_dir: Path) -> Dict[str, Any]:
    """从目录读取并校验 manifest.json。"""
    manifest_path = module_dir / "manifest.json"
    if not manifest_path.exists():
        raise PlatformException(
            code=30004, message="模块清单文件不存在", status_code=400
        )

    try:
        raw = json.loads(manifest_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise PlatformException(
            code=30004, message=f"模块清单文件格式错误：{exc}", status_code=400
        ) from exc

    try:
        manifest = Manifest(**raw)
    except ValidationError as exc:
        raise PlatformException(
            code=30004,
            message=f"模块清单校验失败：{exc.errors()}",
            status_code=400,
        ) from exc

    _validate_manifest(manifest, module_dir=module_dir)
    return raw


def validate_manifest_dict(
    raw: Dict[str, Any], module_dir: Path | None = None
) -> Manifest:
    """校验 manifest 字典并返回 Manifest 对象。"""
    try:
        manifest = Manifest(**raw)
    except ValidationError as exc:
        raise PlatformException(
            code=30004,
            message=f"模块清单校验失败：{exc.errors()}",
            status_code=400,
        ) from exc

    _validate_manifest(manifest, module_dir=module_dir)
    return manifest
