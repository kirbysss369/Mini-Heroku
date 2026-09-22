from __future__ import annotations

from datetime import datetime
from enum import StrEnum
from typing import Any
from uuid import UUID, uuid4

from sqlalchemy import JSON, DateTime, Enum, ForeignKey, Integer, String, Text, func, text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db import Base


class NodeStatus(StrEnum):
    ONLINE = "ONLINE"
    UNREACHABLE = "UNREACHABLE"


class DeploymentStatus(StrEnum):
    PENDING = "PENDING"
    SCHEDULED = "SCHEDULED"
    PULLING = "PULLING"
    STARTING = "STARTING"
    HEALTH_CHECKING = "HEALTH_CHECKING"
    READY = "READY"
    FAILED = "FAILED"


class CommandStatus(StrEnum):
    PENDING = "PENDING"
    RUNNING = "RUNNING"
    SUCCEEDED = "SUCCEEDED"
    FAILED = "FAILED"


class Node(Base):
    __tablename__ = "nodes"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    name: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)
    public_host: Mapped[str] = mapped_column(String(255), nullable=False)
    status: Mapped[NodeStatus] = mapped_column(
        Enum(NodeStatus, name="node_status", native_enum=False, create_constraint=True),
        nullable=False,
        default=NodeStatus.ONLINE,
        server_default=NodeStatus.ONLINE.value,
    )
    last_heartbeat_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=text("CURRENT_TIMESTAMP")
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        server_default=text("CURRENT_TIMESTAMP"),
        onupdate=func.now(),
    )

    deployments: Mapped[list[Deployment]] = relationship(back_populates="node")
    commands: Mapped[list[Command]] = relationship(back_populates="node")


class Application(Base):
    __tablename__ = "applications"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    name: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=text("CURRENT_TIMESTAMP")
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        server_default=text("CURRENT_TIMESTAMP"),
        onupdate=func.now(),
    )

    deployments: Mapped[list[Deployment]] = relationship(back_populates="application")


class Deployment(Base):
    __tablename__ = "deployments"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    application_id: Mapped[UUID] = mapped_column(
        ForeignKey("applications.id"), nullable=False
    )
    node_id: Mapped[UUID | None] = mapped_column(ForeignKey("nodes.id"))
    image: Mapped[str] = mapped_column(String(512), nullable=False)
    container_port: Mapped[int] = mapped_column(Integer, nullable=False)
    host_port: Mapped[int | None] = mapped_column(Integer)
    container_id: Mapped[str | None] = mapped_column(String(255))
    status: Mapped[DeploymentStatus] = mapped_column(
        Enum(
            DeploymentStatus,
            name="deployment_status",
            native_enum=False,
            create_constraint=True,
        ),
        nullable=False,
        default=DeploymentStatus.PENDING,
        server_default=DeploymentStatus.PENDING.value,
    )
    error: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=text("CURRENT_TIMESTAMP")
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        server_default=text("CURRENT_TIMESTAMP"),
        onupdate=func.now(),
    )

    application: Mapped[Application] = relationship(back_populates="deployments")
    node: Mapped[Node | None] = relationship(back_populates="deployments")
    commands: Mapped[list[Command]] = relationship(back_populates="deployment")


class Command(Base):
    __tablename__ = "commands"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    node_id: Mapped[UUID] = mapped_column(ForeignKey("nodes.id"), nullable=False)
    deployment_id: Mapped[UUID] = mapped_column(
        ForeignKey("deployments.id"), nullable=False
    )
    type: Mapped[str] = mapped_column(String(64), nullable=False)
    payload: Mapped[dict[str, Any]] = mapped_column(JSON, nullable=False)
    status: Mapped[CommandStatus] = mapped_column(
        Enum(CommandStatus, name="command_status", native_enum=False, create_constraint=True),
        nullable=False,
        default=CommandStatus.PENDING,
        server_default=CommandStatus.PENDING.value,
    )
    result: Mapped[dict[str, Any] | None] = mapped_column(JSON)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=text("CURRENT_TIMESTAMP")
    )
    started_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    completed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))

    node: Mapped[Node] = relationship(back_populates="commands")
    deployment: Mapped[Deployment] = relationship(back_populates="commands")
