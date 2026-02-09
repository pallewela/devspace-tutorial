# Chapter 2: Your First DevSpace Project

## Learning Objectives

By the end of this chapter, you will:
- Clone and initialize a DevSpace project using `devspace init`
- Understand what files DevSpace creates
- Deploy your first application to Kubernetes with `devspace deploy`
- Verify the app is running in your cluster

## Prerequisites

- Completed [Chapter 1: Introduction and Environment Setup](../01-setup/README.md)
- DevSpace CLI, Docker, and kubectl installed
- A running Kubernetes cluster with a namespace selected

## Choose a Project

For this chapter, we'll use the official DevSpace Go quickstart. This is a minimal Go web server that's perfect for learning DevSpace.

### Clone the Quickstart

```bash
git clone https://github.com/loft-sh/devspace-quickstart-golang
cd devspace-quickstart-golang
```

**What's in this repo?**

Let's look at the important files:

```bash
ls -la
```

You'll see:
- `main.go` — A simple Go HTTP server (listens on port 8080)
- `Dockerfile` — Instructions to build the Go app as a container image
- `go.mod` and `go.sum` — Go dependency files

**Quick look at main.go:**
```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from DevSpace!")
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server listening on :8080")
    http.ListenAndServe(":8080", nil)
}
```

It's a basic web server. When you visit it, it returns "Hello from DevSpace!".

**Quick look at Dockerfile:**
```dockerfile
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
```

This is a multi-stage Dockerfile: build the Go binary in the first stage, then copy it to a minimal Alpine image for smaller size.

## Initialize DevSpace

Now let's tell DevSpace how to deploy this app.

### Run `devspace init`

From the `devspace-quickstart-golang` directory:

```bash
devspace init
```

DevSpace will ask you a series of questions. Let's walk through them:

### Question 1: Programming Language

```
? Select the programming language of this project
  > go
    javascript
    python
    ...
```

**Select: `go`**

This tells DevSpace you're using Go. DevSpace will configure the right dev image (a pre-built image with Go tools for development).

### Question 2: Deployment Method

```
? How do you want to deploy this project?
  > helm
    kubectl
    kustomize
```

**Select: `helm`**

Helm is Kubernetes's package manager. DevSpace uses a built-in "component chart" (a simple, generic Helm chart) to deploy your app. You can also use `kubectl` (raw manifests) or `kustomize`. For simplicity, we'll use Helm.

### Question 3: Is This a Quickstart?

```
? Is this a DevSpace Quickstart project?
  > Yes
    No
```

**Select: `Yes`**

DevSpace auto-detects the official quickstart repos and can configure them automatically.

### Question 4: Development or Deploy Only?

```
? Do you want to develop this project with DevSpace or just deploy it?
  > I want to develop this project and my current working dir contains the source code
    I just want to deploy this project
```

**Select: `I want to develop this project and my current working dir contains the source code`**

This tells DevSpace to set up file sync, port forwarding, and a dev terminal (we'll use these in Chapter 4).

### Question 5: Dockerfile Location

```
? How should DevSpace build the container image for this project?
  > Use this existing Dockerfile: ./Dockerfile
    Use a different Dockerfile
    Skip / I don't know
```

**Select: `Use this existing Dockerfile: ./Dockerfile`**

DevSpace will use the Dockerfile in the repo to build the image.

### Question 6: Container Registry

```
? If you were to push any images, which container registry would you want to push to?
  > Skip Registry
    Use hub.docker.com
    Use GitHub image registry
    Use other registry
```

**Select: `Skip Registry`**

For local development with minikube or kind, you don't need to push images to a registry. DevSpace can build images directly in the cluster or use the local Docker daemon.

### Initialization Complete

```
✓ Project successfully initialized

ℹ Configuration saved in devspace.yaml - you can make adjustments as needed

You can now run:
1. devspace use namespace - to pick which Kubernetes namespace to work in
2. devspace dev - to start developing your project in Kubernetes
```

Great! DevSpace created some files. Let's see what changed.

## What DevSpace Created

Run:

```bash
ls -la
```

**New files:**

1. **`devspace.yaml`** — The DevSpace configuration file. This is the single source of truth for how to build, deploy, and develop your app.
2. **`devspace_start.sh`** — A script that runs when you open a terminal to your dev container. It shows helpful info.
3. **`.devspace/` in `.gitignore`** — DevSpace uses `.devspace/` to cache build state and variables. This folder shouldn't be committed to git.

Let's briefly look at `devspace.yaml`:

```bash
cat devspace.yaml
```

You'll see sections like `pipelines`, `images`, `deployments`, and `dev`. Don't worry about understanding everything yet—we'll dive deep in Chapter 3. For now, know that this file tells DevSpace:
- **What images to build** (`images` section)
- **How to deploy** (`deployments` section)
- **What to do in dev mode** (`dev` section)

## Deploy Your Application

Now for the exciting part: let's deploy the app to Kubernetes!

### Run `devspace deploy`

```bash
devspace deploy
```

**What happens:**

1. **DevSpace runs the `deploy` pipeline** (defined in `devspace.yaml`):
   - Check and deploy any dependencies (none in this project)
   - Build the image from the Dockerfile
   - Tag the image
   - Deploy the app using Helm

You'll see output like:

```
info Using namespace 'devspace-tutorial'
info Using kube context 'minikube'
build:images Executing pipeline 'build' for images
build:image app Building image 'ghcr.io/loft-sh/devspace-quickstart-golang:xxxxxx' with engine 'docker'
build:image app Building with Docker...
...
deploy Deploying with helm...
deploy:app Helm: Release "app" does not exist. Installing it now.
deploy:app Deployed helm chart (Release revision: 1)
✓ Successfully deployed!
```

**Key things that happened:**

1. **Image built**: DevSpace built the Docker image using the `Dockerfile`.
2. **Image tagged**: DevSpace tagged it (you'll see a hash like `xxxxxx`).
3. **Helm chart deployed**: DevSpace deployed a Helm release called `app` with your container.

### Verify the Deployment

Let's check that the app is running in Kubernetes:

```bash
kubectl get pods -n devspace-tutorial
```

You should see:

```
NAME                   READY   STATUS    RESTARTS   AGE
app-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
```

The pod is running! The `1/1` under `READY` means 1 out of 1 containers in the pod is ready.

**Check the deployment:**

```bash
kubectl get deployments -n devspace-tutorial
```

```
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
app    1/1     1            1           1m
```

**Check the service:**

```bash
kubectl get services -n devspace-tutorial
```

```
NAME   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
app    ClusterIP   10.96.123.456   <none>        8080/TCP   1m
```

A Kubernetes Service was created to expose your app on port 8080 (inside the cluster).

## Access Your Application

Your app is running in the cluster, but how do you access it?

### Option 1: Port Forwarding with kubectl

```bash
kubectl port-forward -n devspace-tutorial deployment/app 8080:8080
```

Now open your browser and go to:

```
http://localhost:8080
```

You should see:

```
Hello from DevSpace!
```

Press `Ctrl+C` to stop port forwarding.

### Option 2: Using DevSpace (Preview of Chapter 4)

DevSpace can handle port forwarding automatically in dev mode. We'll cover this in detail in Chapter 4. For now, just know that `devspace dev` sets up port forwarding, file sync, and more.

## Understanding What Happened

Let's recap the `devspace deploy` flow:

```
devspace deploy
     │
     ├─> 1. Run dependencies (none here)
     │
     ├─> 2. Build images
     │       └─> Build 'app' image from Dockerfile
     │       └─> Tag it (e.g., ghcr.io/.../app:abc123)
     │
     └─> 3. Deploy with Helm
             └─> Create Deployment (runs your container as a pod)
             └─> Create Service (exposes port 8080)
```

All of this is defined in `devspace.yaml`. The `deploy` pipeline (in the `pipelines` section) orchestrates these steps.

## Redeploying

Want to redeploy? Just run `devspace deploy` again. DevSpace is smart:
- It only rebuilds the image if files changed
- It updates the Helm release if the config changed

Try it:

```bash
devspace deploy
```

If nothing changed, you'll see:

```
info Skip building image 'app': No changes detected
deploy Deploying with helm...
deploy:app Helm: Release "app" already exists. Upgrading it now.
✓ Successfully deployed!
```

DevSpace skipped the build because nothing changed.

## Clean Up (Optional)

Want to remove the app from your cluster?

```bash
devspace purge
```

This runs the `purge` pipeline, which deletes the Helm release and stops any running dev sessions. You'll see:

```
purge Purging deployment 'app'...
✓ Successfully purged!
```

The app is now gone from your cluster. You can verify:

```bash
kubectl get pods -n devspace-tutorial
```

```
No resources found in devspace-tutorial namespace.
```

**For this tutorial, you can leave the app deployed** (we'll use it in Chapters 3 and 4). If you did purge it, just run `devspace deploy` again later.

## What You Learned

- The `devspace init` wizard asks questions and generates a `devspace.yaml` config file
- `devspace.yaml` is the single source of truth for build, deploy, and dev workflows
- `devspace deploy` builds your image and deploys it to Kubernetes using Helm
- You can verify deployments with kubectl (`kubectl get pods`, `kubectl get deployments`)
- You can access the app via port forwarding (`kubectl port-forward` or DevSpace UI in dev mode)

## Next Steps

Now that you've deployed an app, let's understand how it works. Move on to [Chapter 3: Understanding devspace.yaml](../03-understanding-config/README.md) to learn the anatomy of the config file and how to customize it.

## Troubleshooting

**Problem: Build fails with "docker daemon not running"**
- Make sure Docker is running: `docker ps`
- For minikube: You can build images inside minikube's Docker daemon: `eval $(minikube docker-env)` (then re-run `devspace deploy`)

**Problem: Deployment stuck in "Pending" or "ImagePullBackOff"**
- For local clusters (minikube/kind), make sure you selected "Skip Registry" during init
- If you used a registry, make sure the image was pushed and the cluster can pull it

**Problem: Can't access localhost:8080**
- Make sure port forwarding is running (`kubectl port-forward ...`)
- Check if something else is using port 8080: `lsof -i :8080` (macOS/Linux) or `netstat -ano | findstr :8080` (Windows)
- Try a different local port: `kubectl port-forward -n devspace-tutorial deployment/app 3000:8080` (then visit localhost:3000)

**Problem: `devspace deploy` fails with permission errors**
- Make sure your kube-context has permissions to create resources
- For minikube/kind, you're the admin by default
- For cloud clusters, check your RBAC permissions
