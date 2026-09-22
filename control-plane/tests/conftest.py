import os
from pathlib import Path

import pytest
from alembic import command
from alembic.config import Config
from sqlalchemy import create_engine
from sqlalchemy.orm import Session

from app.models import Application, Command, Deployment, Node

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
