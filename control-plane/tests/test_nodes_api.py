from datetime import UTC, datetime, timedelta
from uuid import UUID, uuid4

from app.models import Node, NodeStatus
from app.services.nodes import mark_stale_nodes_unreachable


def test_first_registration_creates_node(client, db_session) -> None:
    response = client.post(
        "/internal/nodes/register",
        json={"name": "worker-1", "public_host": "127.0.0.1"},
    )

    assert response.status_code == 200
    body = response.json()
    assert body["name"] == "worker-1"
    assert body["public_host"] == "127.0.0.1"
    assert body["status"] == "ONLINE"
    assert body["last_heartbeat_at"] is not None

    node = db_session.get(Node, UUID(body["id"]))
    assert node is not None
    assert node.name == "worker-1"


def test_repeated_registration_updates_existing_node(client, db_session) -> None:
    first = client.post(
        "/internal/nodes/register",
        json={"name": "worker-repeat", "public_host": "127.0.0.1"},
    )
    node_id = first.json()["id"]
    first_heartbeat = datetime.fromisoformat(first.json()["last_heartbeat_at"])

    second = client.post(
        "/internal/nodes/register",
        json={"name": "worker-repeat", "public_host": "10.0.0.8"},
    )

    assert second.status_code == 200
    body = second.json()
    assert body["id"] == node_id
    assert body["public_host"] == "10.0.0.8"
    assert body["status"] == "ONLINE"
    assert datetime.fromisoformat(body["last_heartbeat_at"]) >= first_heartbeat

    nodes = db_session.query(Node).filter(Node.name == "worker-repeat").all()
    assert len(nodes) == 1


def test_heartbeat_marks_node_online(client, db_session) -> None:
    node = Node(
        name="worker-heartbeat",
        public_host="127.0.0.2",
        status=NodeStatus.UNREACHABLE,
        last_heartbeat_at=datetime.now(UTC) - timedelta(minutes=5),
    )
    db_session.add(node)
    db_session.commit()
    old_heartbeat = node.last_heartbeat_at

    response = client.post(f"/internal/nodes/{node.id}/heartbeat")

    assert response.status_code == 200
    db_session.refresh(node)
    assert node.status is NodeStatus.ONLINE
    assert node.last_heartbeat_at is not None
    assert node.last_heartbeat_at > old_heartbeat


def test_heartbeat_unknown_node_returns_404(client) -> None:
    response = client.post(f"/internal/nodes/{uuid4()}/heartbeat")

    assert response.status_code == 404
    assert response.json() == {"detail": "node not found"}


def test_stale_node_detection(db_session) -> None:
    now = datetime.now(UTC)
    stale = Node(
        name="worker-stale",
        public_host="127.0.0.3",
        status=NodeStatus.ONLINE,
        last_heartbeat_at=now - timedelta(seconds=61),
    )
    fresh = Node(
        name="worker-fresh",
        public_host="127.0.0.4",
        status=NodeStatus.ONLINE,
        last_heartbeat_at=now - timedelta(seconds=5),
    )
    db_session.add_all([stale, fresh])
    db_session.commit()

    changed = mark_stale_nodes_unreachable(
        db_session,
        now=now,
    )

    db_session.refresh(stale)
    db_session.refresh(fresh)
    assert changed == 1
    assert stale.status is NodeStatus.UNREACHABLE
    assert fresh.status is NodeStatus.ONLINE
