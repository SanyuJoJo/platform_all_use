from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import List
import logging
logger = logging.getLogger(__name__)
class Settings(BaseSettings):
    """应用配置，从环境变量加载"""
    APP_NAME: str = "Platform Backend"
    APP_ENV: str = "development"
    DEBUG: bool = True
    SECRET_KEY: str = "change-this-in-production"
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 1440
    DATABASE_URL: str = "sqlite+aiosqlite:///./app.db"
    LOG_LEVEL: str = "INFO"
    # CORS 允许的源，逗号分隔；若留空或仅空白，则自动回退为 "*"
    CORS_ORIGINS: str = "*"
    @property
    def cors_origins_list(self) -> List[str]:
        """
        将逗号分隔的字符串转换为列表。
        若配置为空或仅空白，则返回 ["*"] 确保跨域请求不被阻断。
        """
        raw = self.CORS_ORIGINS
        if not raw or raw.strip() == "":
            logger.warning("CORS_ORIGINS 未配置或为空，已自动回退为 ['*']，生产环境请显式指定域名")
            return ["*"]
        return [origin.strip() for origin in raw.split(",") if origin.strip()]
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
    )
settings = Settings()
