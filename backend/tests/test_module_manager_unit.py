"""模块管理模块 - 纯函数单元测试（v1.3，V12-P1-10）。"""
from pathlib import Path

from src.modules.module_manager.manifest import _resolve_entry_file
from src.modules.module_manager.service import filter_menus_by_permission


def test_filter_menus_recursive_parent_denied():
    """父菜单无权限 → 整棵子树跳过。"""
    menus = [
        {"id": "m:a", "parent_id": None, "permission": "m:a:view"},
        {"id": "m:b", "parent_id": "m:a", "permission": "m:b:view"},
        {"id": "m:c", "parent_id": None, "permission": "m:c:view"},
    ]
    result = filter_menus_by_permission(menus, ["m:c:view"])
    ids = [m["id"] for m in result]
    assert ids == ["m:c"]


def test_filter_menus_recursive_child_denied():
    """父菜单有权限、子菜单无权限 → 仅子菜单跳过。"""
    menus = [
        {"id": "m:a", "parent_id": None, "permission": "m:a:view"},
        {"id": "m:b", "parent_id": "m:a", "permission": "m:b:view"},
    ]
    result = filter_menus_by_permission(menus, ["m:a:view"])
    ids = [m["id"] for m in result]
    assert "m:a" in ids
    assert "m:b" not in ids


def test_filter_menus_orphan_treated_as_root():
    """parent_id 引用不存在父 → 视为根节点。"""
    menus = [
        {"id": "m:x", "parent_id": "not_exist", "permission": "m:x:view"},
    ]
    result = filter_menus_by_permission(menus, ["m:x:view"])
    assert [m["id"] for m in result] == ["m:x"]


def test_resolve_entry_file_simple(tmp_path: Path):
    (tmp_path / "router.py").write_text("")
    assert _resolve_entry_file(tmp_path, "router") is not None


def test_resolve_entry_file_dotted(tmp_path: Path):
    (tmp_path / "api").mkdir()
    (tmp_path / "api" / "router.py").write_text("")
    assert _resolve_entry_file(tmp_path, "api.router") is not None


def test_resolve_entry_file_package(tmp_path: Path):
    (tmp_path / "api").mkdir()
    (tmp_path / "api" / "__init__.py").write_text("")
    assert _resolve_entry_file(tmp_path, "api") is not None


def test_resolve_entry_file_reject_py_suffix(tmp_path: Path):
    """V12-P1-04：显式拒绝 .py 后缀。"""
    (tmp_path / "router.py").write_text("")
    assert _resolve_entry_file(tmp_path, "router.py") is None


def test_resolve_entry_file_missing(tmp_path: Path):
    assert _resolve_entry_file(tmp_path, "router") is None
