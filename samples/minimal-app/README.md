# Minimal Go App for DevSpace Tutorial

A minimal Go web server for practicing DevSpace configuration (used in Chapters 3, 5, and 6).

## Features

- Simple HTTP server on configurable port
- Health check endpoint (`/health`)
- Environment variable support
- Multi-stage Dockerfile for small image size

## Usage

### Run Locally (Without DevSpace)

```bash
go run main.go
```

Open `http://localhost:8080` in your browser.

### Run with DevSpace

```bash
devspace dev
```

Edit `main.go` locally. DevSpace syncs changes to the container. Rebuild and restart the app inside the container:

```bash
go run main.go
```

### Deploy

```bash
devspace deploy
```

### Access the App

```bash
kubectl port-forward -n devspace-tutorial deployment/app 8080:8080
```

Open `http://localhost:8080`.

## Environment Variables

- `PORT`: Server port (default: `8080`)
- `ENVIRONMENT`: Environment name (default: `development`)

## Endpoints

- `/`: Returns a greeting with the environment name
- `/health`: Health check endpoint (returns `OK`)
