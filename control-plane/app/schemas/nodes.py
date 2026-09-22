from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict

from app.models import NodeStatus


class NodeRegisterRequest(BaseModel):
    name: str
    public_host: str


class NodeResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: UUID
    name: str
    public_host: str
    status: NodeStatus
    last_heartbeat_at: datetime | None
