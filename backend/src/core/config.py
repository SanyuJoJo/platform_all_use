from pydantic_settings import BaseSettings, SettingsConfigDict
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
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
    )
settings = Settings()
