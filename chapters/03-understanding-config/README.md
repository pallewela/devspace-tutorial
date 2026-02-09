# Chapter 3: Understanding devspace.yaml

## Learning Objectives

By the end of this chapter, you will:
- Understand the structure of `devspace.yaml`
- Know what each major section does (pipelines, images, deployments, dev)
- Be able to make small, safe changes to the config
- Understand how pipelines orchestrate build and deploy steps

## Prerequisites

- Completed [Chapter 2: Your First DevSpace Project](../02-first-project/README.md)
- Have the `devspace-quickstart-golang` project initialized

## Overview

The `devspace.yaml` file is the **single source of truth** for how DevSpace builds, deploys, and develops your application. Think of it like a `Dockerfile` (defines how to build an image) combined with Kubernetes manifests (defines what to deploy) plus development settings (file sync, ports, terminal).

Let's open the file and understand each section.

## Open devspace.yaml

From the `devspace-quickstart-golang` directory:

```bash
cat devspace.yaml
```

Or open it in your favorite editor. The file has this structure:

```yaml
version: v2beta1
name: devspace-quickstart-golang

pipelines:
  dev: ...
  deploy: ...

images:
  app: ...

deployments:
  app: ...

dev:
  app: ...
```

Let's break down each section.

## Top-Level Fields

### `version`

```yaml
version: v2beta1
```

The DevSpace config schema version. Always use `v2beta1` for DevSpace 6.x.

### `name`

```yaml
name: devspace-quickstart-golang
```

The project name. DevSpace uses this internally (e.g., in logs and as a label on deployed resources).

## Pipelines Section

Pipelines define **what happens when you run DevSpace commands**.

```yaml
pipelines:
  dev:
    run: |-
      run_dependencies --all
      create_deployments --all
      start_dev app
  
  deploy:
    run: |-
      run_dependencies --all
      build_images --all -t $(git describe --always)
      create_deployments --all
```

Think of pipelines as **scripts** that call special DevSpace functions.

### `dev` Pipeline

Runs when you execute `devspace dev`:

```yaml
dev:
  run: |-
    run_dependencies --all       # 1. Deploy any project dependencies
    create_deployments --all     # 2. Deploy this project
    start_dev app                # 3. Start dev mode for 'app'
```

**What each line does:**

1. **`run_dependencies --all`**: If your project depends on other services (defined in a `dependencies` section—more in Chapter 8), deploy them first.
2. **`create_deployments --all`**: Deploy all entries in the `deployments` section (in this case, just `app`).
3. **`start_dev app`**: Start the dev configuration named `app` (from the `dev` section). This starts file sync, port forwarding, and opens a terminal.

Notice: The `dev` pipeline **does not build images** by default. Why? In dev mode, DevSpace replaces your production image with a `devImage` (a pre-built image with dev tools). This skips the slow build step during development. (More in Chapter 4.)

### `deploy` Pipeline

Runs when you execute `devspace deploy`:

```yaml
deploy:
  run: |-
    run_dependencies --all
    build_images --all -t $(git describe --always)
    create_deployments --all
```

**What each line does:**

1. **`run_dependencies --all`**: Deploy dependencies (if any).
2. **`build_images --all -t $(git describe --always)`**: Build all images in the `images` section and tag them with the git commit hash (e.g., `abc1234`).
3. **`create_deployments --all`**: Deploy all entries in the `deployments` section.

This is your production deployment pipeline. It builds images, tags them, and deploys.

### Pipeline Syntax

Pipelines use **POSIX shell syntax**, but DevSpace emulates it cross-platform (works the same on Linux, macOS, and Windows). You can use:
- Built-in functions: `build_images`, `create_deployments`, `start_dev`, etc.
- Shell features: `if`, `||`, `&&`, variables, etc.
- Commands: `$(git describe --always)` runs git and captures output

We'll explore custom pipelines in Chapter 7.

## Images Section

The `images` section tells DevSpace **which container images to build**.

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    dockerfile: ./Dockerfile
```

**Fields:**

- **Key (`app`)**: The name of this image. You reference it elsewhere (e.g., in deployments or dev config).
- **`image`**: The full image name (registry + repo + name). This is where the image would be pushed if you used a registry.
- **`dockerfile`**: Path to the Dockerfile (relative to the project root).

### How Images Are Used

1. **During `devspace deploy`**: The `build_images` function in the `deploy` pipeline builds this image.
2. **In deployments**: The deployment references this image (we'll see how below).
3. **During `devspace dev`**: The image is replaced with a `devImage` (defined in the `dev` section), so you skip building.

### Optional Fields

You can also specify:
- **`context`**: Build context path (defaults to `.`)
- **`buildArgs`**: Docker build arguments (e.g., `NODE_VERSION: "18"`)
- **`target`**: Multi-stage Dockerfile target
- **`rebuildStrategy`**: When to rebuild (default, always, or ignoreContextChanges)

We'll explore these in Chapter 5.

## Deployments Section

The `deployments` section tells DevSpace **how to deploy your app** to Kubernetes.

```yaml
deployments:
  app:
    helm:
      chart:
        name: component-chart
        repo: https://charts.devspace.sh
      values:
        containers:
        - image: ghcr.io/loft-sh/devspace-quickstart-golang
        service:
          ports:
          - port: 8080
```

**Fields:**

- **Key (`app`)**: The name of this deployment. You reference it in pipelines (e.g., `create_deployments app`).
- **`helm`**: Use Helm to deploy. (You can also use `kubectl` for raw YAML or `kustomize`.)

### Helm Deployment

DevSpace uses Helm to deploy your app. Here's what each part means:

#### `chart`

```yaml
chart:
  name: component-chart
  repo: https://charts.devspace.sh
```

- **`component-chart`**: A generic, built-in Helm chart provided by DevSpace. It's designed for simple microservices (containers, services, ingress, etc.).
- **`repo`**: Where to fetch the chart from (DevSpace's chart repository).

You don't need to write Helm templates yourself. The component-chart is flexible and covers most use cases.

#### `values`

```yaml
values:
  containers:
  - image: ghcr.io/loft-sh/devspace-quickstart-golang
  service:
    ports:
    - port: 8080
```

These are Helm chart values. They tell the chart:
- **`containers`**: Run a container with this image.
- **`service.ports`**: Expose port 8080 (create a Kubernetes Service).

When you run `devspace deploy`, DevSpace passes these values to Helm, which creates:
- A **Deployment** (manages the pods running your container)
- A **Service** (exposes port 8080 inside the cluster)

### Where Does the Image Tag Come From?

Notice the `image` field doesn't include a tag (no `:v1.0` or `:abc123`). DevSpace automatically injects the tag based on the image it built. So when you run `devspace deploy`:

1. DevSpace builds `ghcr.io/loft-sh/devspace-quickstart-golang` and tags it (e.g., `abc123`).
2. DevSpace replaces `image: ghcr.io/.../golang` with `image: ghcr.io/.../golang:abc123` in the values.
3. Helm deploys with the correct tag.

This ensures your deployment always uses the image you just built.

## Dev Section

The `dev` section defines **what happens during `devspace dev`**.

```yaml
dev:
  app:
    imageSelector: ghcr.io/loft-sh/devspace-quickstart-golang
    devImage: ghcr.io/loft-sh/devspace-containers/go:1.21-alpine
    ports:
    - port: "2345"
    - port: "8080"
    open:
    - url: http://localhost:8080
    terminal:
      command: ./devspace_start.sh
    sync:
    - path: ./
    ssh:
      enabled: true
    proxyCommands:
    - command: devspace
    - command: kubectl
    - command: helm
    - command: git
```

**Key (`app`)**: The name of this dev config. You reference it in the `dev` pipeline (`start_dev app`).

Let's break down each field.

### `imageSelector`

```yaml
imageSelector: ghcr.io/loft-sh/devspace-quickstart-golang
```

Which container to target. DevSpace finds the pod running this image and connects to it.

### `devImage`

```yaml
devImage: ghcr.io/loft-sh/devspace-containers/go:1.21-alpine
```

**This is key**: Instead of using your production image (built from `Dockerfile`), DevSpace **replaces** it with this pre-built dev image.

**Why?**
- Your Dockerfile produces a minimal production image (just the binary, no Go compiler, no dev tools).
- The `devImage` includes the Go toolchain, so you can build and run your app **inside the container** while developing.
- You skip rebuilding the image every time you change code. DevSpace syncs your code and rebuilds it in the container.

This is how DevSpace achieves **hot reloading** without rebuilding Docker images.

### `ports`

```yaml
ports:
- port: "2345"
- port: "8080"
```

Port forwarding. DevSpace forwards these ports from the container to your localhost:
- **`2345`**: Go debugger port (Delve). If you attach a debugger, it connects here.
- **`8080`**: The app's HTTP port.

You can access the app at `http://localhost:8080` while `devspace dev` is running.

### `open`

```yaml
open:
- url: http://localhost:8080
```

DevSpace automatically opens this URL in your browser once the container is ready. Convenient!

### `terminal`

```yaml
terminal:
  command: ./devspace_start.sh
```

DevSpace opens an interactive terminal to the container and runs this command. `devspace_start.sh` is a script created by `devspace init`. It prints helpful info (like "Run `go run main.go` to start the app").

### `sync`

```yaml
sync:
- path: ./
```

File synchronization. DevSpace watches your local files (`./` = everything in the project) and syncs changes to the container in real time.

**How it works:**
1. You edit `main.go` locally.
2. DevSpace detects the change and uploads `main.go` to the container (same path).
3. If you're running the app with a hot-reload tool (like `nodemon` for Node.js or `air` for Go), it restarts automatically.

This is **bi-directional**: If you create a file in the container, it syncs back to your local machine.

### `ssh`

```yaml
ssh:
  enabled: true
```

DevSpace injects a lightweight SSH server into the container. Your IDE (like VS Code with the Remote-SSH extension) can connect to it. This enables "remote development" inside the container.

### `proxyCommands`

```yaml
proxyCommands:
- command: devspace
- command: kubectl
- command: helm
- command: git
```

When you're in the dev container's terminal, these commands are proxied to your **local machine**. So when you run `kubectl get pods` inside the container, it actually runs on your laptop (using your local kubectl config).

This is super useful: you stay in the container context but can still use your local tools.

## Hands-On: Make a Change

Let's modify the config and see what happens.

### Change the Port

Open `devspace.yaml` and change the HTTP port from `8080` to `3000`:

1. In the `deployments.app.helm.values.service.ports` section:

```yaml
service:
  ports:
  - port: 3000  # Changed from 8080
```

2. In the `dev.app.ports` section:

```yaml
ports:
- port: "2345"
- port: "3000:8080"  # Forward localhost:3000 to container:8080
```

3. In the `dev.app.open` section:

```yaml
open:
- url: http://localhost:3000
```

### Also update main.go (Optional)

If you want the app to listen on 3000 (instead of 8080), edit `main.go`:

```go
func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server listening on :3000")
    http.ListenAndServe(":3000", nil)  // Changed from :8080
}
```

And update the Dockerfile `EXPOSE` line:

```dockerfile
EXPOSE 3000
```

### Redeploy

```bash
devspace deploy
```

DevSpace will:
1. Rebuild the image (because `Dockerfile` and `main.go` changed—if you edited them)
2. Upgrade the Helm release with the new port

### Verify

```bash
kubectl port-forward -n devspace-tutorial deployment/app 3000:3000
```

Open `http://localhost:3000` in your browser. You should see "Hello from DevSpace!".

**Revert the changes if you want** (or keep them—it's your project now!).

## Viewing the Full Config

Want to see what DevSpace sees after processing variables and defaults?

```bash
devspace print
```

This prints the fully resolved config. Useful for debugging.

## What You Learned

- `devspace.yaml` has five main sections: `version`, `name`, `pipelines`, `images`, `deployments`, `dev`
- **Pipelines** orchestrate build/deploy/dev workflows using special functions (`build_images`, `create_deployments`, `start_dev`)
- **Images** define which Docker images to build (Dockerfile path, image name)
- **Deployments** define how to deploy to Kubernetes (Helm, kubectl, or kustomize)
- **Dev** defines the development experience (file sync, port forwarding, terminal, devImage)
- You can customize the config and redeploy with `devspace deploy`

## Next Steps

Now that you understand the config, let's use it for development. Move on to [Chapter 4: Development Mode (devspace dev)](../04-development-mode/README.md) to experience hot reloading, file sync, and the DevSpace UI.

## Troubleshooting

**Problem: `devspace print` shows errors**
- Your YAML syntax might be invalid. Check indentation (use spaces, not tabs).
- Run `devspace print` to see the exact error.

**Problem: Changes don't take effect after `devspace deploy`**
- Make sure you saved the file.
- Run `devspace purge` then `devspace deploy` to force a clean deployment.

**Problem: Don't understand a config option**
- Check the [official DevSpace config reference](https://devspace.sh/docs/configuration/reference).
