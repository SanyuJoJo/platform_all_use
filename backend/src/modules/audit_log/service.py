from datetime import datetime, timezone, timedelta
from typing import Optional, List
import csv
import io
from src.core.exceptions import PlatformException

# ========== 模拟日志数据（扩充至 30 条） ==========
_now = datetime.now(timezone.utc)
BASE_TIME = _now - timedelta(days=7)  # 最近7天

MOCK_LOGS = [
    # ---- 认证模块日志 ----
    {
        "id": 1,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "login",
        "resource": "user",
        "resource_id": "1",
        "detail": "管理员登录系统",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-001",
        "created_at": BASE_TIME + timedelta(hours=1)
    },
    {
        "id": 2,
        "user_id": 2,
        "username": "guest",
        "module_id": "auth",
        "action": "login",
        "resource": "user",
        "resource_id": "2",
        "detail": "访客登录系统",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "success",
        "error_code": None,
        "request_id": "req-002",
        "created_at": BASE_TIME + timedelta(hours=2)
    },
    {
        "id": 3,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "logout",
        "resource": "user",
        "resource_id": "1",
        "detail": "管理员登出",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-003",
        "created_at": BASE_TIME + timedelta(hours=3)
    },
    # ---- 用户管理模块日志 ----
    {
        "id": 4,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "create",
        "resource": "user",
        "resource_id": "3",
        "detail": "创建用户 zhangsan",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-004",
        "created_at": BASE_TIME + timedelta(hours=4)
    },
    {
        "id": 5,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "update",
        "resource": "user",
        "resource_id": "3",
        "detail": "更新用户 zhangsan 的角色",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-005",
        "created_at": BASE_TIME + timedelta(hours=5)
    },
    {
        "id": 6,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "delete",
        "resource": "user",
        "resource_id": "4",
        "detail": "删除用户 lisi（软删除）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-006",
        "created_at": BASE_TIME + timedelta(hours=6)
    },
    # ---- 角色管理模块日志 ----
    {
        "id": 7,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "create",
        "resource": "role",
        "resource_id": "3",
        "detail": "创建角色 项目管理员",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-007",
        "created_at": BASE_TIME + timedelta(hours=7)
    },
    {
        "id": 8,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "update",
        "resource": "role",
        "resource_id": "2",
        "detail": "修改访客角色权限",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-008",
        "created_at": BASE_TIME + timedelta(hours=8)
    },
    {
        "id": 9,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "delete",
        "resource": "role",
        "resource_id": "4",
        "detail": "删除角色 临时角色（未使用）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "fail",
        "error_code": 20003,
        "request_id": "req-009",
        "created_at": BASE_TIME + timedelta(hours=9)
    },
    # ---- 模块管理模块日志 ----
    {
        "id": 10,
        "user_id": 1,
        "username": "admin",
        "module_id": "module_manager",
        "action": "create",
        "resource": "module",
        "resource_id": "new_module_abc",
        "detail": "安装新模块 new_module_abc",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-010",
        "created_at": BASE_TIME + timedelta(hours=10)
    },
    {
        "id": 11,
        "user_id": 1,
        "username": "admin",
        "module_id": "module_manager",
        "action": "update",
        "resource": "module",
        "resource_id": "customer_relation",
        "detail": "启用模块 客户关系管理",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-011",
        "created_at": BASE_TIME + timedelta(hours=11)
    },
    {
        "id": 12,
        "user_id": 1,
        "username": "admin",
        "module_id": "module_manager",
        "action": "delete",
        "resource": "module",
        "resource_id": "old_module_xyz",
        "detail": "卸载模块 old_module_xyz（强制）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-012",
        "created_at": BASE_TIME + timedelta(hours=12)
    },
    # ---- 日志审计模块自身操作 ----
    {
        "id": 13,
        "user_id": 1,
        "username": "admin",
        "module_id": "audit_log",
        "action": "view",
        "resource": "log",
        "resource_id": None,
        "detail": "查看日志列表（查询条件：module=auth）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-013",
        "created_at": BASE_TIME + timedelta(hours=13)
    },
    {
        "id": 14,
        "user_id": 1,
        "username": "admin",
        "module_id": "audit_log",
        "action": "export",
        "resource": "log",
        "resource_id": None,
        "detail": "导出日志 CSV",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-014",
        "created_at": BASE_TIME + timedelta(hours=14)
    },
    # ---- 业务模块（客户管理）日志 ----
    {
        "id": 15,
        "user_id": 2,
        "username": "guest",
        "module_id": "customer_relation",
        "action": "view",
        "resource": "customer",
        "resource_id": "101",
        "detail": "查看客户 张三",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "success",
        "error_code": None,
        "request_id": "req-015",
        "created_at": BASE_TIME + timedelta(hours=15)
    },
    {
        "id": 16,
        "user_id": 1,
        "username": "admin",
        "module_id": "customer_relation",
        "action": "create",
        "resource": "customer",
        "resource_id": "102",
        "detail": "创建客户 李四",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-016",
        "created_at": BASE_TIME + timedelta(hours=16)
    },
    {
        "id": 17,
        "user_id": 1,
        "username": "admin",
        "module_id": "customer_relation",
        "action": "update",
        "resource": "customer",
        "resource_id": "102",
        "detail": "更新客户 李四 的联系方式",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-017",
        "created_at": BASE_TIME + timedelta(hours=17)
    },
    {
        "id": 18,
        "user_id": 1,
        "username": "admin",
        "module_id": "customer_relation",
        "action": "delete",
        "resource": "customer",
        "resource_id": "103",
        "detail": "删除客户 王五（客户流失）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-018",
        "created_at": BASE_TIME + timedelta(hours=18)
    },
    # ---- 项目管理模块日志 ----
    {
        "id": 19,
        "user_id": 2,
        "username": "guest",
        "module_id": "project_management",
        "action": "view",
        "resource": "project",
        "resource_id": "P001",
        "detail": "查看项目 平台重构",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "success",
        "error_code": None,
        "request_id": "req-019",
        "created_at": BASE_TIME + timedelta(hours=19)
    },
    {
        "id": 20,
        "user_id": 1,
        "username": "admin",
        "module_id": "project_management",
        "action": "create",
        "resource": "project",
        "resource_id": "P002",
        "detail": "创建项目 微服务改造",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-020",
        "created_at": BASE_TIME + timedelta(hours=20)
    },
    # ---- License 模块日志 ----
    {
        "id": 21,
        "user_id": 1,
        "username": "admin",
        "module_id": "license",
        "action": "create",
        "resource": "license",
        "resource_id": "LIC-001",
        "detail": "导入 License 文件 enterprise.lic",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-021",
        "created_at": BASE_TIME + timedelta(hours=21)
    },
    {
        "id": 22,
        "user_id": 1,
        "username": "admin",
        "module_id": "license",
        "action": "view",
        "resource": "license",
        "resource_id": None,
        "detail": "查看 License 状态",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-022",
        "created_at": BASE_TIME + timedelta(hours=22)
    },
    # ---- 失败操作日志 ----
    {
        "id": 23,
        "user_id": 2,
        "username": "guest",
        "module_id": "auth",
        "action": "update",
        "resource": "user",
        "resource_id": "2",
        "detail": "尝试修改自己的密码（旧密码错误）",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "fail",
        "error_code": 10003,
        "request_id": "req-023",
        "created_at": BASE_TIME + timedelta(hours=23)
    },
    {
        "id": 24,
        "user_id": 2,
        "username": "guest",
        "module_id": "auth",
        "action": "view",
        "resource": "user",
        "resource_id": "1",
        "detail": "访客尝试查看管理员信息（权限不足）",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "fail",
        "error_code": 20051,
        "request_id": "req-024",
        "created_at": BASE_TIME + timedelta(hours=24)
    },
    # ---- 更多随机日志 ----
    {
        "id": 25,
        "user_id": 1,
        "username": "admin",
        "module_id": "module_manager",
        "action": "update",
        "resource": "module",
        "resource_id": "module-manager",
        "detail": "更新模块管理配置",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-025",
        "created_at": BASE_TIME + timedelta(hours=25)
    },
    {
        "id": 26,
        "user_id": 1,
        "username": "admin",
        "module_id": "auth",
        "action": "login",
        "resource": "user",
        "resource_id": "1",
        "detail": "管理员登录（来自新设备）",
        "ip": "10.0.0.5",
        "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)",
        "status": "success",
        "error_code": None,
        "request_id": "req-026",
        "created_at": BASE_TIME + timedelta(hours=26)
    },
    {
        "id": 27,
        "user_id": None,
        "username": None,
        "module_id": "auth",
        "action": "login",
        "resource": "user",
        "resource_id": None,
        "detail": "登录失败：用户名不存在（attempt: hacker）",
        "ip": "203.0.113.1",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "fail",
        "error_code": 10001,
        "request_id": "req-027",
        "created_at": BASE_TIME + timedelta(hours=27)
    },
    {
        "id": 28,
        "user_id": 1,
        "username": "admin",
        "module_id": "customer_relation",
        "action": "export",
        "resource": "customer",
        "resource_id": None,
        "detail": "导出客户数据（Excel）",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-028",
        "created_at": BASE_TIME + timedelta(hours=28)
    },
    {
        "id": 29,
        "user_id": 2,
        "username": "guest",
        "module_id": "project_management",
        "action": "update",
        "resource": "project",
        "resource_id": "P001",
        "detail": "访客尝试更新项目状态（权限不足）",
        "ip": "192.168.1.101",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "status": "fail",
        "error_code": 20051,
        "request_id": "req-029",
        "created_at": BASE_TIME + timedelta(hours=29)
    },
    {
        "id": 30,
        "user_id": 1,
        "username": "admin",
        "module_id": "audit_log",
        "action": "view",
        "resource": "log",
        "resource_id": None,
        "detail": "查看全部日志列表",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        "status": "success",
        "error_code": None,
        "request_id": "req-030",
        "created_at": BASE_TIME + timedelta(hours=30)
    },
]

_next_log_id = 31

# ========== 核心业务函数 ==========

def list_logs(
    page: int,
    page_size: int,
    module_id: Optional[str] = None,
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    start_time: Optional[str] = None,
    end_time: Optional[str] = None,
    keyword: Optional[str] = None,
) -> dict:
    logs = MOCK_LOGS
    if module_id:
        logs = [l for l in logs if l["module_id"] == module_id]
    if user_id:
        logs = [l for l in logs if l["user_id"] == user_id]
    if action:
        logs = [l for l in logs if l["action"] == action]
    if keyword:
        keyword = keyword.lower()
        logs = [l for l in logs if keyword in (l.get("detail") or "").lower() or keyword in (l.get("username") or "").lower()]
    if start_time:
        logs = [l for l in logs if l["created_at"].isoformat() >= start_time]
    if end_time:
        logs = [l for l in logs if l["created_at"].isoformat() <= end_time]
    total = len(logs)
    start = (page - 1) * page_size
    end = start + page_size
    items = logs[start:end]
    return {
        "items": items,
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }

def get_log_by_id(log_id: int) -> Optional[dict]:
    for log in MOCK_LOGS:
        if log["id"] == log_id:
            return log
    return None

def export_logs_csv(filters: dict) -> str:
    logs = list_logs(
        page=1,
        page_size=10000,
        module_id=filters.get("module_id"),
        user_id=filters.get("user_id"),
        action=filters.get("action"),
        start_time=filters.get("start_time"),
        end_time=filters.get("end_time"),
        keyword=filters.get("keyword"),
    )
    items = logs["items"]
    if not items:
        return ""
    output = io.StringIO()
    fieldnames = list(items[0].keys())
    writer = csv.DictWriter(output, fieldnames=fieldnames)
    writer.writeheader()
    for row in items:
        row_copy = row.copy()
        row_copy["created_at"] = row_copy["created_at"].isoformat()
        writer.writerow(row_copy)
    return output.getvalue()
