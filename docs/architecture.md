# Mini PaaS v0.1 Architecture

## Purpose

Mini PaaS is a portfolio-grade system for learning and demonstrating the complete lifecycle of deploying a containerized application. It is inspired by Heroku, but v0.1 intentionally favors explicit, understandable components over production-scale infrastructure.

The intended v0.1 lifecycle is:

```text
create app
-> create deployment
-> choose a healthy worker node
-> enqueue deployment command
-> Go agent receives the command
-> agent starts a Docker container
-> agent performs healthcheck
-> agent reports result
-> deployment becomes READY or FAILED
```

The control plane now has the core persistence model for that lifecycle, but scheduling and deployment execution are not implemented yet.

## Components

### Control Plane

The control plane is a Python 3.12+ FastAPI service. PostgreSQL is the source of truth for applications, deployments, worker nodes, and commands.

The current control plane contains:
- a minimal FastAPI application;
- `GET /health`;
- environment-based settings;
- SQLAlchemy 2.x models and session configuration;
- Alembic migrations;
- pytest coverage for the health endpoint and database model behavior.

No scheduling, command dispatch, or deployment execution logic exists yet.

### Worker Agent

The worker agent is written in Go. The architecture uses a pull model: the control plane never initiates a connection to an agent.

In the completed v0.1 design, each agent will:
1. register itself;
2. send heartbeats periodically;
3. poll the control plane for commands;
4. execute Docker operations;
5. report command results.

The bootstrap only loads configuration, installs structured logging, waits for process signals, and provides a graceful-shutdown hook. It does not contact the control plane or Docker.

### Demo App

The demo app exposes `GET /health` and `GET /` with version information and is independently buildable as a Docker image.

### PostgreSQL

PostgreSQL is the only infrastructure service in the root Compose file and remains the source of truth for v0.1.

## Planned v0.1 state machines

Deployment states:

```text
PENDING
SCHEDULED
PULLING
STARTING
HEALTH_CHECKING
READY
FAILED
```

Worker-node states:

```text
ONLINE
UNREACHABLE
```

## Architectural constraints

- Agents use a pull model.
- PostgreSQL is the source of truth.
- Business logic belongs outside FastAPI route handlers.
- Docker-specific agent logic will live behind a Go interface when introduced.
- Simple implementations are preferred over speculative abstractions.
- Important non-trivial business logic must have tests.
- Errors must retain enough context for debugging and must not be swallowed.

## Out of scope for v0.1

Unless explicitly requested by a later milestone:
- frontend/UI;
- user authentication and authorization;
- Kubernetes;
- Redis;
- Celery;
- RabbitMQ;
- Kafka;
- external message brokers;
- autoscaling;
- high availability;
- distributed consensus;
- production secrets management;
- TLS termination and custom domains;
- billing;
- production observability stacks.

Scheduling, agent registration, heartbeats, command polling, Docker Engine integration, health-check execution, and deployment state transitions remain intentionally postponed until later milestones.
