"""应用配置。

v1.4 变更：
- 新增 HOST / PORT / WORKERS / FRONTEND_DEPLOY_DIR 字段的显式声明。
  这些字段由安装脚本写入 .env，但后端 Settings 直接使用 uvicorn / os.getenv
  读取，无需在此消费。声明它们可避免 Pydantic v2 的 ValidationError。
- model_config 增加 extra="ignore"，容忍 .env 中未来可能出现的新字段。
"""
import logging

from pydantic_settings import BaseSettings, SettingsConfigDict

logger = logging.getLogger(__name__)


class Settings(BaseSettings):
    """应用配置，从环境变量加载。"""

    # ---- 应用信息 ----
    APP_NAME: str = "Platform Backend"
    APP_ENV: str = "development"
    DEBUG: bool = True
    LOG_LEVEL: str = "INFO"

    # ---- 安全 ----
    SECRET_KEY: str = "change-this-in-production"
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 1440

    # ---- 数据库 ----
    DATABASE_URL: str = "sqlite+aiosqlite:///./app.db"

    # ---- 模块管理 ----
    MODULES_DIR: str = "src/modules"
    MODULE_UPLOAD_DIR: str = "./uploads/modules"
    MODULE_ZIP_MAX_SIZE: int = 50 * 1024 * 1024       # 单文件 50MB
    MODULE_ZIP_MAX_TOTAL: int = 200 * 1024 * 1024     # 解压总大小 200MB
    MODULE_ZIP_MAX_FILES: int = 2000                  # 文件数量上限

    # ---- License 管理 ----
    LICENSE_SECRET_KEY: str = "change-this-license-secret-in-production"
    LICENSE_ACTIVATION_URL: str = ""
    LICENSE_MACHINE_CODE_OVERRIDE: str = ""

    # ---- 前端静态资源 ----
    # 由后端 main.py 通过 os.getenv("FRONTEND_DEPLOY_DIR") 读取。
    # 在 Settings 中显式声明，仅为容忍 .env 中存在该字段（避免 ValidationError）。
    FRONTEND_DEPLOY_DIR: str = ""

    # ---- 部署脚本相关（uvicorn 启动参数） ----
    # 由 install.sh 写入 .env，再由 start.sh / systemd 通过 source / EnvironmentFile
    # 读取并传给 uvicorn。后端进程自身不直接使用这三个字段；
    # 在此声明仅为容忍 .env 中存在它们。
    HOST: str = "0.0.0.0"
    PORT: int = 8000
    WORKERS: int = 1

    # ---- CORS ----
    CORS_ORIGINS: str = (
        "http://localhost:3000,http://localhost:5173,"
        "http://127.0.0.1:3000,http://127.0.0.1:5173"
    )

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
        # ★ 关键：忽略 .env 中未声明的字段，
        #   避免安装脚本新增字段时后端启动失败。
        extra="ignore",
    )

    @property
    def cors_origins_list(self) -> list[str]:
        """将逗号分隔的 CORS_ORIGINS 转换为列表。"""
        raw = (self.CORS_ORIGINS or "").strip()
        if not raw:
            if self.APP_ENV.lower() == "production":
                logger.error(
                    "生产环境未配置 CORS_ORIGINS，已回退为 ['*']，"
                    "这是不安全配置，请立即在环境变量中指定允许的域名"
                )
            else:
                logger.warning(
                    "CORS_ORIGINS 未配置，回退为 ['*']，"
                    "生产环境请显式指定允许的域名"
                )
            return ["*"]
        if raw == "*":
            return ["*"]
        return [item.strip() for item in raw.split(",") if item.strip()]


settings = Settings()
