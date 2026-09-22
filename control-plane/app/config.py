from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "Mini PaaS Control Plane"
    environment: str = "development"
    database_url: str = (
        "postgresql+psycopg://mini_paas:mini_paas@localhost:5432/mini_paas"
    )
    node_stale_after_seconds: int = Field(default=30, gt=0)

    model_config = SettingsConfigDict(
        env_prefix="MINIPAAS_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )


@lru_cache
def get_settings() -> Settings:
    return Settings()
