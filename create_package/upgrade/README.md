# 系统升级功能（v1.9）

平台新增升级能力：**动作尽量小（只重启平台后端，无关机）、可回退、支持推送升级提醒**。

## 升级原理

```
发布方                                   目标机
build_upgrade_script.sh   ──>  <INSTALL_DIR>/update/<包名>-<版本>/
  ├─ backend.tar.gz / dist.war           └─ manifest.json
  └─ manifest.json
                                       ┌──────────────────────────────┐
GET /api/v1/upgrade/status             │  后端扫描 update/ 目录          │
  → 检测到 to_version > current         │  标记 update_available=true   │
POST /api/v1/upgrade/run               │  → 编排 apply.sh               │
  → 槽位切换 + restart.sh              │  （只重启后端，约 1-2 秒）       │
POST /api/v1/upgrade/rollback          │  → 编排 rollback.sh            │
  → 恢复上一槽位 + restart.sh          │  （不删数据）                   │
                                       └──────────────────────────────┘
```

## 一、发布方：生成升级包

```bash
cd create_package/upgrade

# 1. 先构建完整载荷（backend.tar.gz / dist.war）
bash ../../build-product.sh          # 或分别跑 backend-go/scripts/build.sh --flat

# 2. 生成升级包（可选 --product backend|frontend|core|all）
export ED25519_PRIVATE_KEY="<离线私钥>"     # 可选；不签名则仅做 sha 校验
bash build_upgrade_script.sh --product all \
  --from 1.6.0 --to 1.7.0 --channel prod \
  --out /path/to/release

# 产物：/path/to/release/platform-upgrade-1.7.0/
```

## 二、目标机：部署 + 升级

```bash
# 1. 把升级包目录放到安装根的 update/ 下
scp -r platform-upgrade-1.7.0 <target>:/opt/platstool/update/

# 2. （可选）体检
/opt/platstool/bin/upgrade/preflight.sh --install-dir /opt/platstool --pkg platform-upgrade-1.7.0

# 3. 通过管理界面执行升级，或直接调用脚本：
/opt/platstool/bin/upgrade/apply.sh --pkg platform-upgrade-1.7.0 --install-dir /opt/platstool
# 升级成功输出回滚 id：rb-20261127-...-platform-upgrade-1.7.0
```

## 三、升级动作说明（为什么小）

- `apply.sh` 只做**槽位切换**（`backend/`、`frontend/deploy/` 各自 `.new → .prev → 原子 mv`），**无关机、无 reboot**。
- 仅通过 `bin/restart.sh`（stop + start）重启平台后端，停机窗口约 1–2 秒。
- **保留**：`backend/data/`、`.env`、密钥、`core/data/`（数据不触碰）。

## 四、回退

```bash
# 方式一：管理界面「系统升级 → 回退」输入回退 id
# 方式二：脚本
/opt/platstool/bin/upgrade/rollback.sh --id rb-20261127-... --install-dir /opt/platstool
```
回退从安装根之外的 `<install>/.plat_rollback/<id>/` 恢复上一槽位，**不删除数据**。

## 五、推送升级提醒

管理界面「系统升级」页：
- 顶部展示当前版本 + `可用更新` 状态（后端扫描 `update/` 得到）。
- 有可用更新时出现黄色提醒条。
- 执行升级/回退均需二次确认，且仅管理员角色可操作（后端强制 `admin/super_admin` + 权限码）。

后端 API：
- `GET  /api/v1/upgrade/status`   → 当前版本 + 可用升级清单（提醒数据源）
- `GET  /api/v1/upgrade/packages` → 包列表（含校验状态）
- `POST /api/v1/upgrade/run`      → 执行升级（body: `{"package_id":"..."}`）
- `POST /api/v1/upgrade/rollback` → 回退（body: `{"id":"..."}`）

## 六、权限

| 路由 | 权限码 | 角色 |
|---|---|---|
| `GET /upgrade/status` | `upgrade:package:view` | 授权即可 |
| `GET /upgrade/packages` | `upgrade:package:view` | 授权即可 |
| `POST /upgrade/run` | `upgrade:package:run` | **admin/super_admin** |
| `POST /upgrade/rollback` | `upgrade:package:rollback` | **admin/super_admin** |

升级写接口**不**进入超管直通，且强制管理员角色——降权后的旧 token 也无法触发升级。