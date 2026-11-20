# nshttps_gm 统一入口代理（国密 HTTPS，自包含）

`nshttps_gm` 是基于 Tengine（集成铜锁 Tongsuo + NTLS 国密模块）的**自包含**统一前端代理，
**内部集成了 Tongsuo 源码**（不依赖外部第三方库），为平台所有对外服务提供统一的 **8443 端口 HTTPS 入口**，
支持**国际 TLS** 与**国密 NTLS** 双协议共存，自动生成 SM2 国密证书（10 年有效），
并可**一键打包部署**到任意 x86_64 Linux 服务器。

## 目录结构

```
nshttps_gm/
├── tongsuo/             # 内部集成的 Tongsuo（铜锁）源码（自包含，不依赖外部库）
├── tengine-3.1.0/       # Tengine 源码 + 已编译 nginx（含 NTLS 模块）
├── tongsuo-bin/         # 构建产物：内部 Tongsuo 运行库（openssl + libcrypto/libssl）
├── build.sh             # 自包含构建：编译内部 Tongsuo + Tengine
├── gen-certs.sh         # 自动生成国密证书（SM2 + ECC，3650 天）
├── gen-config.sh        # 生成 nginx 配置（8443 双协议共存，反代内部服务）
├── build-package.sh     # 打包一键部署包
├── install.sh           # 部署包内的一键部署脚本
├── start.sh / stop.sh / status.sh
├── out/                 # 生成的部署包（gitignore）
└── README.md
```

## 自包含构建（编译）

```bash
cd nshttps_gm
# 内部集成了 tongsuo 源码，编译过程无需外部依赖（只需 gcc/make）
bash build.sh            # 编译内部 Tongsuo + Tengine，产物 rpath 指向内部 tongsuo
```

构建后：
- `tongsuo-bin/` 内含 `bin/openssl`、`libs/libcrypto.so.3`、`libs/libssl.so.3`、`ssl/openssl.cnf`
- `tengine-3.1.0/bin/sbin/nginx` 的 rpath 指向 `tongsuo-bin/libs`（完全自包含）

## 一键打包部署

```bash
# 1) 生成部署包（在开发机/编译机）
cd nshttps_gm
bash build-package.sh
# 输出：out/nshttps-gm-<版本>-<时间戳>.tar.gz（约 15M，含运行产物，不含源码）

# 2) 部署到目标服务器（x86_64 Linux）
tar -xzf nshttps-gm-1.0.0-*.tar.gz
bash install.sh --prefix /opt/nshttps_gm --port 8443 --upstream 127.0.0.1:8000

# 3) 部署目录内初始化并启动
cd /opt/nshttps_gm
bash gen-certs.sh --cn <你的域名>          # 自动签发 10 年国密+ECC 证书
bash gen-config.sh --port 8443 --upstream 127.0.0.1:8000
bash start.sh
```

部署目录 `/opt/nshttps_gm` 结构：
```
/opt/nshttps_gm/
├── tengine-3.1.0/       # 已编译 nginx
├── tongsuo-bin/         # 内部 Tongsuo 运行库
├── gen-certs.sh  gen-config.sh  start.sh  stop.sh  status.sh
└── conf/  certs/  logs/  pid/  tmp/  run/
```

## 生成证书说明

`gen-certs.sh` 默认有效期 **3650 天（10 年）**，一次生成长期有效，避免频繁过期：
- **SM2 根 CA** + **SM2 签名证书** + **SM2 加密证书** → 国密 NTLS 通道
- **独立 ECC(P-256) 根 CA** + **ECC 国际证书** → 国际 TLS 通道（标准浏览器可用）
- 注：SM2 CA 只能以 SM3 签名，无法签 SHA-256 的 ECC 证书，故国际通道用独立 ECC CA

输出到 `certs/`：
```
certs/ca/sm2-root.crt         SM2 根 CA（国密）
certs/ca/ecc-root.crt         ECC(P-256) 根 CA（国际）
certs/server/server.crt        SM2 服务器签名证书
certs/server/server_enc.crt    SM2 服务器加密证书
certs/server/international.crt ECC 服务器证书（国际）
```

## 统一入口设计

- 对外只暴露 **一个端口 8443**，同时支持：
  - **国际 TLS**（标准浏览器/客户端）：用 ECC 证书，TLSv1.2/1.3
  - **国密 NTLS**（国密浏览器/铜锁）：`enable_ntls on` + SM2 双证书，套件 `ECC-SM2-SM4-*`
- **单向认证**：服务端出示证书，不强制客户端证书（双向可放开注释）
- **反代内部服务**：默认 `127.0.0.1:8000`（`--upstream` 可改），内部服务统一走回环

## 端口说明

- 默认监听 `8443`（`gen-config.sh --port` 或 `install.sh --port` 可改）
- 安全组需放行 `8443/tcp`
- 反代上游默认 `127.0.0.1:8000`（`--upstream` 可改）

## 国密 NTLS 套件

- `ECC-SM2-SM4-CBC-SM3`、`ECC-SM2-SM4-GCM-SM3`（静态，实测可用）
- `ECDHE-SM2-SM4-CBC-SM3`、`ECDHE-SM2-SM4-GCM-SM3`（动态，铜锁 s_client 单测偶发内部错误，属客户端侧限制，不影响服务端）

## 依赖

- 编译时：gcc、make、perl（Tongsuo/Tengine 编译需要）
- 运行时：无外部依赖（内部集成 tongsuo 运行库）
- 目标服务器：x86_64 Linux（内核 ≥ 3.x，glibc ≥ 2.17）