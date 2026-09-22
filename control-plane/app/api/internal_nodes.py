from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.db import get_db
from app.models import Node
from app.schemas.nodes import NodeRegisterRequest, NodeResponse
from app.services.nodes import NodeNotFoundError, record_heartbeat, register_node

router = APIRouter(prefix="/internal/nodes", tags=["internal-nodes"])
DbSession = Annotated[Session, Depends(get_db)]


@router.post("/register", response_model=NodeResponse)
def register(payload: NodeRegisterRequest, session: DbSession) -> Node:
    return register_node(
        session,
        name=payload.name,
        public_host=payload.public_host,
    )


@router.post("/{node_id}/heartbeat", response_model=NodeResponse)
def heartbeat(node_id: UUID, session: DbSession) -> Node:
    try:
        return record_heartbeat(session, node_id)
    except NodeNotFoundError as exc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="node not found",
        ) from exc
