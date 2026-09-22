from uuid import UUID

import pytest
from sqlalchemy.exc import IntegrityError

from app.models import (
    Application,
    Command,
    CommandStatus,
    Deployment,
    DeploymentStatus,
    Node,
    NodeStatus,
)


def test_application_creation(db_session) -> None:
    application = Application(name="demo-app")
    db_session.add(application)
    db_session.commit()

    assert isinstance(application.id, UUID)
    assert application.name == "demo-app"
    assert application.created_at is not None
    assert application.updated_at is not None


def test_node_creation(db_session) -> None:
    node = Node(name="worker-1", public_host="worker-1.local")
    db_session.add(node)
    db_session.commit()

    assert isinstance(node.id, UUID)
    assert node.status is NodeStatus.ONLINE
    assert node.last_heartbeat_at is None


def test_deployment_creation_and_relationships(db_session) -> None:
    application = Application(name="api")
    node = Node(name="worker-2", public_host="worker-2.local")
    deployment = Deployment(
        application=application,
        node=node,
        image="example/demo:0.1.0",
        container_port=8080,
    )
    db_session.add(deployment)
    db_session.commit()

    assert deployment.status is DeploymentStatus.PENDING
    assert deployment.application is application
    assert deployment.node is node
    assert deployment in application.deployments
    assert deployment in node.deployments


def test_command_creation_and_relationships(db_session) -> None:
    application = Application(name="worker-api")
    node = Node(name="worker-3", public_host="worker-3.local")
    deployment = Deployment(
        application=application,
        node=node,
        image="example/demo:0.1.0",
        container_port=8080,
    )
    command = Command(
        node=node,
        deployment=deployment,
        type="START_DEPLOYMENT",
        payload={"image": "example/demo:0.1.0"},
    )
    db_session.add(command)
    db_session.commit()

    assert command.status is CommandStatus.PENDING
    assert command.node is node
    assert command.deployment is deployment
    assert command in node.commands
    assert command in deployment.commands


def test_application_name_is_unique(db_session) -> None:
    db_session.add(Application(name="duplicate-app"))
    db_session.commit()
    db_session.add(Application(name="duplicate-app"))

    with pytest.raises(IntegrityError):
        db_session.commit()


def test_node_name_is_unique(db_session) -> None:
    db_session.add(Node(name="duplicate-node", public_host="worker-a.local"))
    db_session.commit()
    db_session.add(Node(name="duplicate-node", public_host="worker-b.local"))

    with pytest.raises(IntegrityError):
        db_session.commit()
