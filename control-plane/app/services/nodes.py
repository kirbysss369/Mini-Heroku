from datetime import UTC, datetime, timedelta
from uuid import UUID

from sqlalchemy import select, update
from sqlalchemy.orm import Session

from app.config import get_settings
from app.models import Node, NodeStatus


class NodeNotFoundError(LookupError):
    def __init__(self, node_id: UUID) -> None:
        super().__init__(f"node {node_id} not found")
        self.node_id = node_id


def utc_now() -> datetime:
    return datetime.now(UTC)


def register_node(
    session: Session,
    *,
    name: str,
    public_host: str,
    now: datetime | None = None,
) -> Node:
    heartbeat_at = now or utc_now()
    node = session.scalar(select(Node).where(Node.name == name))

    if node is None:
        node = Node(
            name=name,
            public_host=public_host,
            status=NodeStatus.ONLINE,
            last_heartbeat_at=heartbeat_at,
        )
        session.add(node)
    else:
        node.public_host = public_host
        node.status = NodeStatus.ONLINE
        node.last_heartbeat_at = heartbeat_at

    session.commit()
    session.refresh(node)
    return node


def record_heartbeat(
    session: Session,
    node_id: UUID,
    *,
    now: datetime | None = None,
) -> Node:
    node = session.get(Node, node_id)
    if node is None:
        raise NodeNotFoundError(node_id)

    node.status = NodeStatus.ONLINE
    node.last_heartbeat_at = now or utc_now()
    session.commit()
    session.refresh(node)
    return node


def mark_stale_nodes_unreachable(
    session: Session,
    *,
    now: datetime | None = None,
    stale_after_seconds: int | None = None,
) -> int:
    threshold_seconds = (
        stale_after_seconds
        if stale_after_seconds is not None
        else get_settings().node_stale_after_seconds
    )
    cutoff = (now or utc_now()) - timedelta(seconds=threshold_seconds)

    result = session.execute(
        update(Node)
        .where(
            Node.status == NodeStatus.ONLINE,
            Node.last_heartbeat_at.is_not(None),
            Node.last_heartbeat_at <= cutoff,
        )
        .values(status=NodeStatus.UNREACHABLE)
    )
    session.commit()
    return result.rowcount or 0
