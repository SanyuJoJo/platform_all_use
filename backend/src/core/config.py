"""应用配置。"""
import logging

from pydantic_settings import BaseSettings, SettingsConfigDict

logger = logging.getLogger(__name__)


class Settings(BaseSettings):
    """应用配置，从环境变量加载。"""

    APP_NAME: str = "Platform Backend"
    APP_ENV: str = "development"
    DEBUG: bool = True
    SECRET_KEY: str = "change-this-in-production"
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 1440
    DATABASE_URL: str = "sqlite+aiosqlite:///./app.db"
    LOG_LEVEL: str = "INFO"

    # 模块源码根目录（相对 backend 工作目录）
    MODULES_DIR: str = "src/modules"
    # 模块上传/安装的受控根目录（source_path 白名单）
    MODULE_UPLOAD_DIR: str = "./uploads/modules"
    # ZIP 安装限制
    MODULE_ZIP_MAX_SIZE: int = 50 * 1024 * 1024       # 单文件 50MB
    MODULE_ZIP_MAX_TOTAL: int = 200 * 1024 * 1024     # 解压总大小 200MB
    MODULE_ZIP_MAX_FILES: int = 2000                  # 文件数量上限

    # CORS
    CORS_ORIGINS: str = (
        "http://localhost:3000,http://localhost:5173,"
        "http://127.0.0.1:3000,http://127.0.0.1:5173"
    )

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
    )

    @property
    def cors_origins_list(self) -> list[str]:
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
