# Multi-Service Example for DevSpace Tutorial

A two-service application demonstrating DevSpace dependencies (used in Chapter 8).

## Structure

```
multi-service/
├── devspace.yaml         # Main config with dependencies
├── api/
│   ├── main.go          # API service (port 8080)
│   ├── go.mod
│   ├── Dockerfile
│   └── devspace.yaml    # API deployment config
└── frontend/
    ├── main.go          # Frontend service (port 3000)
    ├── go.mod
    ├── Dockerfile
    └── devspace.yaml    # Frontend deployment config
```

## Services

### API

- **Port**: 8080
- **Endpoint**: `/api` - Returns JSON: `{"message": "Hello from API!"}`

### Frontend

- **Port**: 3000
- **Endpoint**: `/` - Calls the API and displays the result

## Usage

### Deploy All Services

From the `multi-service/` directory:

```bash
devspace deploy
```

This deploys:
1. API (dependency)
2. Frontend (dependency)
3. Main project (just prints a message)

### Verify

```bash
kubectl get pods -n devspace-tutorial
```

You should see:
- `api-xxxxx`
- `frontend-xxxxx`

### Access the Frontend

```bash
kubectl port-forward -n devspace-tutorial deployment/frontend 3000:3000
```

Open `http://localhost:3000`. You should see:

```
Frontend calling API: {"message": "Hello from API!"}
```

The frontend calls the API internally (using Kubernetes service DNS: `http://api:8080`).

### Dev Mode (All Services)

```bash
devspace dev
```

This starts dev mode for all dependencies. You can edit code in `api/` or `frontend/` and see changes synced.

### Deploy Individual Services

If you want to deploy only the API:

```bash
cd api
devspace deploy
```

Or only the frontend:

```bash
cd frontend
devspace deploy
```

### Clean Up

```bash
devspace purge
```

## How Dependencies Work

The main `devspace.yaml` defines:

```yaml
dependencies:
  api:
    path: ./api
  frontend:
    path: ./frontend
```

When you run `devspace deploy` in the main directory:
1. DevSpace deploys `./api` (runs its `deploy` pipeline)
2. DevSpace deploys `./frontend` (runs its `deploy` pipeline)
3. DevSpace runs the main pipeline (just prints a message)

The frontend can access the API via `http://api:8080` because both are in the same namespace and Kubernetes creates a service DNS entry for each service.
