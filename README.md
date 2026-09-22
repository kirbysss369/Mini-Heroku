# Mini-Heroku

Mini-Heroku is a portfolio-grade Mini PaaS inspired by Heroku. v0.1 focuses on understanding the complete deployment lifecycle, not on building a production PaaS.

This repository currently contains the initial monorepo bootstrap only. Deployment logic is intentionally not implemented yet.

## Repository layout

- `control-plane/` — FastAPI control plane and PostgreSQL access.
- `agent/` — Go worker-agent skeleton.
- `demo-app/` — tiny FastAPI application used by later deployment milestones.
- `docs/` — architecture and ADRs.
- `infra/` — infrastructure files added by later milestones.
- `docker-compose.yml` — local PostgreSQL for development.

## Quick start

Start PostgreSQL:

```bash
make up
```

Install control-plane dependencies and run tests:

```bash
cd control-plane
python3.12 -m venv .venv
. .venv/bin/activate
pip install -e ".[dev]"
pytest
```

Run Go tests:

```bash
cd agent
go test ./...
```

Build the demo app:

```bash
docker build -t mini-paas-demo:0.1.0 demo-app
```

See `docs/architecture.md` for scope and architecture.
