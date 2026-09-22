import os
from pathlib import Path

import pytest
from alembic.config import Config
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import Session

from alembic import command
from app.db import get_db
from app.main import app

CONTROL_PLANE_DIR = Path(__file__).resolve().parents[1]
DEFAULT_DATABASE_URL = "postgresql+psycopg://mini_paas:mini_paas@localhost:5432/mini_paas"


def database_url() -> str:
    return os.getenv("MINIPAAS_TEST_DATABASE_URL") or os.getenv(
        "MINIPAAS_DATABASE_URL", DEFAULT_DATABASE_URL
    )


@pytest.fixture(scope="session")
def db_engine():
    url = database_url()
    alembic_config = Config(str(CONTROL_PLANE_DIR / "alembic.ini"))
    alembic_config.set_main_option("script_location", str(CONTROL_PLANE_DIR / "alembic"))
    alembic_config.attributes["database_url"] = url
    command.upgrade(alembic_config, "head")

    engine = create_engine(url, pool_pre_ping=True)
    yield engine
    engine.dispose()


@pytest.fixture
def db_session(db_engine):
    connection = db_engine.connect()
    transaction = connection.begin()
    session = Session(bind=connection, join_transaction_mode="create_savepoint")

    try:
        yield session
    finally:
        session.close()
        transaction.rollback()
        connection.close()


@pytest.fixture
def client(db_session):
    def override_get_db():
        yield db_session

    app.dependency_overrides[get_db] = override_get_db
    try:
        with TestClient(app) as test_client:
            yield test_client
    finally:
        app.dependency_overrides.clear()
