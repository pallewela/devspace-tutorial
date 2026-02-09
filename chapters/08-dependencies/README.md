# Chapter 8: Dependencies

## Learning Objectives

By the end of this chapter, you will:
- Understand what dependencies are and when to use them
- Define path-based dependencies (monorepo or local folders)
- Define git-based dependencies (separate repositories)
- Reference dependency images in your deployments
- Control execution order of dependencies
- Pass variables to dependencies

## Prerequisites

- Completed [Chapter 7: Pipelines](../07-pipelines/README.md)
- Have the `devspace-quickstart-golang` project initialized

## What Are Dependencies?

In DevSpace, a **dependency** is another project with its own `devspace.yaml`. Dependencies let you:

- **Compose multi-service applications**: Your frontend depends on an API, which depends on a database.
- **Reuse existing configs**: Don't duplicate deployment logic; reference another project's `devspace.yaml`.
- **Work across repositories**: Your app can depend on a service in a different git repo.

When you run `devspace deploy`, DevSpace deploys **dependencies first**, then your project.

### Use Cases

1. **Monorepo**: Multiple services in one repo (e.g., `./api`, `./frontend`, `./worker`). Each has a `devspace.yaml`.
2. **Microservices across repos**: Your app depends on an auth service in `github.com/company/auth-service`.
3. **Shared infrastructure**: Multiple projects depend on a common database or cache service.

## Dependency Sources

DevSpace supports two types of dependencies:

1. **Path-based**: Local folders (e.g., `./api`, `./frontend`)
2. **Git-based**: Remote repositories (e.g., `github.com/company/auth-service`)

## Path-Based Dependencies

Path-based dependencies are perfect for monorepos.

### Example Structure

```
my-monorepo/
├── devspace.yaml        # Main project
├── api/
│   ├── devspace.yaml    # API service
│   └── ...
└── frontend/
    ├── devspace.yaml    # Frontend service
    └── ...
```

### Define the Dependency

In the main `devspace.yaml`:

```yaml
dependencies:
  api:
    path: ./api
  frontend:
    path: ./frontend
```

When you run `devspace deploy` in the main project, DevSpace:
1. Deploys `./api` (runs its `deploy` pipeline)
2. Deploys `./frontend` (runs its `deploy` pipeline)
3. Deploys the main project

### Execution Order

By default, dependencies are deployed **in parallel**. If you need sequential deployment:

```yaml
pipelines:
  deploy:
    run: |-
      create_deployments api --from dependencies
      create_deployments frontend --from dependencies
      create_deployments main
```

Or use `run_dependencies` sequentially:

```yaml
pipelines:
  deploy:
    run: |-
      run_dependencies api
      run_dependencies frontend
      build_images --all
      create_deployments --all
```

## Git-Based Dependencies

Git-based dependencies let you depend on projects in other repositories.

### Example

Your app depends on an auth service:

```yaml
dependencies:
  auth-service:
    git: https://github.com/company/auth-service
    branch: main
```

When you run `devspace deploy`, DevSpace:
1. Clones `github.com/company/auth-service` (caches it locally in `~/.devspace/`)
2. Runs the `deploy` pipeline from `auth-service/devspace.yaml`
3. Deploys your app

### Git Options

You can specify:

- **`branch`**: Git branch (e.g., `main`, `develop`)
- **`tag`**: Git tag (e.g., `v1.2.3`)
- **`revision`**: Git commit hash (e.g., `abc1234`)
- **`subPath`**: Path to the `devspace.yaml` within the repo (e.g., `/services/auth`)

**Example:**

```yaml
dependencies:
  auth-service:
    git: https://github.com/company/auth-service
    tag: v1.0.0
    subPath: /backend
```

DevSpace clones the repo, checks out tag `v1.0.0`, and looks for `devspace.yaml` in `/backend`.

## Hands-On: Create a Two-Service App

Let's create a simple two-service app: an API and a frontend. We'll use path-based dependencies.

### Step 1: Create the Directory Structure

From your tutorial directory:

```bash
cd /home/pallewela/localdev/devspace-tutorial
mkdir -p samples/multi-service/api
mkdir -p samples/multi-service/frontend
cd samples/multi-service
```

### Step 2: Create the API Service

**Create `api/main.go`:**

```bash
cat > api/main.go <<'EOF'
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "{\"message\": \"Hello from API!\"}")
}

func main() {
    http.HandleFunc("/api", handler)
    fmt.Println("API listening on :8080")
    http.ListenAndServe(":8080", nil)
}
EOF
```

**Create `api/go.mod`:**

```bash
cat > api/go.mod <<'EOF'
module api

go 1.21
EOF
```

**Create `api/Dockerfile`:**

```bash
cat > api/Dockerfile <<'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o /app/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080
CMD ["/app/server"]
EOF
```

**Create `api/devspace.yaml`:**

```bash
cat > api/devspace.yaml <<'EOF'
version: v2beta1
name: api

images:
  api:
    image: my-registry/api
    dockerfile: ./Dockerfile

deployments:
  api:
    helm:
      chart:
        name: component-chart
        repo: https://charts.devspace.sh
      values:
        containers:
        - image: my-registry/api
        service:
          ports:
          - port: 8080
EOF
```

### Step 3: Create the Frontend Service

**Create `frontend/main.go`:**

```bash
cat > frontend/main.go <<'EOF'
package main

import (
    "fmt"
    "io"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Call the API
    resp, err := http.Get("http://api:8080/api")
    if err != nil {
        fmt.Fprintf(w, "Error calling API: %v", err)
        return
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Fprintf(w, "Frontend calling API: %s", body)
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Frontend listening on :3000")
    http.ListenAndServe(":3000", nil)
}
EOF
```

**Create `frontend/go.mod`:**

```bash
cat > frontend/go.mod <<'EOF'
module frontend

go 1.21
EOF
```

**Create `frontend/Dockerfile`:**

```bash
cat > frontend/Dockerfile <<'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o /app/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 3000
CMD ["/app/server"]
EOF
```

**Create `frontend/devspace.yaml`:**

```bash
cat > frontend/devspace.yaml <<'EOF'
version: v2beta1
name: frontend

images:
  frontend:
    image: my-registry/frontend
    dockerfile: ./Dockerfile

deployments:
  frontend:
    helm:
      chart:
        name: component-chart
        repo: https://charts.devspace.sh
      values:
        containers:
        - image: my-registry/frontend
        service:
          ports:
          - port: 3000
EOF
```

### Step 4: Create the Main devspace.yaml

Create `samples/multi-service/devspace.yaml`:

```bash
cat > devspace.yaml <<'EOF'
version: v2beta1
name: multi-service

dependencies:
  api:
    path: ./api
  frontend:
    path: ./frontend

pipelines:
  deploy:
    run: |-
      run_dependencies --all
      echo "All services deployed!"
  
  dev:
    run: |-
      run_dependencies --all --pipeline dev
      echo "All services running in dev mode!"
EOF
```

### Step 5: Deploy

From `samples/multi-service`:

```bash
devspace deploy
```

DevSpace:
1. Deploys `api` (builds image, deploys to cluster)
2. Deploys `frontend` (builds image, deploys to cluster)
3. Prints "All services deployed!"

### Step 6: Verify

```bash
kubectl get pods -n devspace-tutorial
```

You should see:

```
NAME                        READY   STATUS    RESTARTS   AGE
api-xxxxxxxxxx-xxxxx        1/1     Running   0          30s
frontend-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
```

### Step 7: Access the Frontend

Port-forward to the frontend:

```bash
kubectl port-forward -n devspace-tutorial deployment/frontend 3000:3000
```

Open `http://localhost:3000` in your browser. You should see:

```
Frontend calling API: {"message": "Hello from API!"}
```

The frontend called the API and returned the result!

## Referencing Dependency Images

You can reference images from dependencies in your deployments or dev config.

### Runtime Variables

DevSpace provides runtime variables for dependency images:

```
${runtime.dependencies.<dep-name>.images.<image-name>.image}
${runtime.dependencies.<dep-name>.images.<image-name>.tag}
```

### Example: Use API Image in Frontend Deployment

If your frontend needs to know the API image (e.g., for a sidecar), you can reference it:

```yaml
dependencies:
  api:
    path: ./api

deployments:
  frontend:
    helm:
      values:
        containers:
        - image: my-registry/frontend
        - image: ${runtime.dependencies.api.images.api.image}:${runtime.dependencies.api.images.api.tag}
          name: api-sidecar
```

DevSpace injects the API image and tag at deploy time.

## Passing Variables to Dependencies

You can pass variables to dependencies via the `vars` field:

```yaml
dependencies:
  api:
    path: ./api
    vars:
      API_PORT: "9090"
      ENVIRONMENT: "staging"
```

The dependency's `devspace.yaml` can use these:

```yaml
# In api/devspace.yaml
vars:
  API_PORT: "8080"  # Default
  ENVIRONMENT: "dev"

deployments:
  api:
    helm:
      values:
        service:
          ports:
          - port: ${API_PORT}
```

When deployed as a dependency, `API_PORT` is `9090` (overridden by the parent).

## Deploying to Different Namespaces

You can deploy dependencies to different namespaces:

```yaml
dependencies:
  api:
    path: ./api
    namespace: api-namespace
  frontend:
    path: ./frontend
    namespace: frontend-namespace
```

DevSpace creates the namespaces if they don't exist.

## Ignoring Nested Dependencies

If a dependency has its own dependencies, you can skip them:

```yaml
dependencies:
  api:
    path: ./api
    ignoreDependencies: true
```

Only the API is deployed; its dependencies are ignored.

## What You Learned

- **Dependencies** let you compose multi-service apps by referencing other `devspace.yaml` configs
- **Path-based dependencies** are for monorepos or local folders
- **Git-based dependencies** are for services in other repositories
- Dependencies are deployed **before** the main project (in parallel by default)
- You can **reference dependency images** using runtime variables
- You can **pass variables** to dependencies to customize their config
- Dependencies can be deployed to **different namespaces**

## Next Steps

You now know how to work with multi-service apps. Let's dive deeper into dev containers and IDE integration. Move on to [Chapter 9: Dev Containers and IDE Integration](../09-dev-containers-ide/README.md).

## Troubleshooting

**Problem: Dependency not deploying**
- Make sure the path is correct: `ls ./api/devspace.yaml`
- Check the dependency's `devspace.yaml` for errors: `cd api && devspace print`
- Run with verbose logs: `devspace deploy --debug`

**Problem: Dependency deployed but not accessible**
- Check if the service is running: `kubectl get pods -n <namespace>`
- Verify the service name matches what you're calling (e.g., `http://api:8080`)
- Check if the dependency is in the same namespace (or use a fully-qualified DNS name: `http://api.api-namespace.svc.cluster.local:8080`)

**Problem: Git dependency fails to clone**
- Make sure the git URL is correct and accessible
- For private repos, ensure your SSH keys or HTTPS credentials are set up
- DevSpace caches clones in `~/.devspace/`. To re-clone: `rm -rf ~/.devspace/git/...`

**Problem: Circular dependency detected**
- DevSpace detects if A depends on B and B depends on A
- Restructure your dependencies to avoid cycles (e.g., extract shared logic to a third dependency)
