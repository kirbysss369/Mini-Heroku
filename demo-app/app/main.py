from fastapi import FastAPI

APP_VERSION = "0.1.0"

app = FastAPI(title="Mini PaaS Demo App", version=APP_VERSION)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/")
def root() -> dict[str, str]:
    return {
        "service": "mini-paas-demo-app",
        "version": APP_VERSION,
    }
