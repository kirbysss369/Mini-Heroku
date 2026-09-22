PYTHON ?= python3.12

.PHONY: up down test test-python test-go lint

up:
	docker compose up -d postgres

down:
	docker compose down

test: test-python test-go

test-python:
	cd control-plane && $(PYTHON) -m pytest

test-go:
	cd agent && go test ./...

lint:
	cd control-plane && $(PYTHON) -m ruff check .
	@unformatted="$$(gofmt -l agent)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt required for:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	cd agent && go vet ./...
