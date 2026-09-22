# Deployment lifecycle

## Deployment states

- `PENDING` — The deployment exists but has not been assigned to a worker node.
- `SCHEDULED` — A worker node has been selected and deployment work can be queued.
- `PULLING` — The worker is pulling the requested container image.
- `STARTING` — The worker is creating and starting the container.
- `HEALTH_CHECKING` — The container is running and its health endpoint is being checked.
- `READY` — The deployment passed its health check and is available for use.
- `FAILED` — The deployment could not complete and its error field describes the failure.

## Command states

- `PENDING` — The command is waiting for its worker node to claim it.
- `RUNNING` — The worker has claimed the command and is executing it.
- `SUCCEEDED` — The command completed successfully.
- `FAILED` — The command finished with an error.
