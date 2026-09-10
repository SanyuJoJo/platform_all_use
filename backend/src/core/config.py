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
    # CORS 允许的源，逗号分隔；设为 "*" 表示允许所有（此时自动关闭 credentials）
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
        """
        将 CORS_ORIGINS 字符串解析为列表。
        - 空值：回退为 ["*"]，并视 APP_ENV 记录 WARNING / ERROR 日志（N-4）；
        - "*"：直接返回 ["*"]；
        - 其他：按逗号切分并去除空白项。
        """
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
