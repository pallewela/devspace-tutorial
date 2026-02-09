# Chapter 5: Images and Builds

## Learning Objectives

By the end of this chapter, you will:
- Understand the `images` section of `devspace.yaml` in depth
- Know when and how images are built
- Use build arguments to customize image builds
- Understand image tagging and how tags flow to deployments
- Configure registry settings and pull secrets

## Prerequisites

- Completed [Chapter 4: Development Mode (devspace dev)](../04-development-mode/README.md)
- Have the `devspace-quickstart-golang` project initialized

## The Images Section

The `images` section tells DevSpace which container images to build, how to build them, and where to push them.

Here's the config from the quickstart:

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    dockerfile: ./Dockerfile
```

Let's explore all available options.

## Image Configuration Options

### Required Fields

#### `image`

```yaml
image: myregistry.com/myrepo/myapp
```

The full image name (registry + repository + image name). This is where the image would be pushed.

**Examples:**
- Docker Hub: `myusername/myapp`
- GitHub Container Registry: `ghcr.io/myusername/myapp`
- Google Container Registry: `gcr.io/my-project/myapp`
- Private registry: `registry.company.com/team/myapp`

### Optional Fields

#### `dockerfile`

```yaml
dockerfile: ./Dockerfile
```

Path to the Dockerfile (relative to the project root). Defaults to `./Dockerfile`.

**Example:** If your Dockerfile is in a subdirectory:

```yaml
dockerfile: ./docker/app.Dockerfile
```

#### `context`

```yaml
context: ./
```

The Docker build context (the directory sent to the Docker daemon). Defaults to `.` (project root).

**Example:** If your code is in a `src/` subdirectory:

```yaml
context: ./src
```

#### `buildArgs`

```yaml
buildArgs:
  GO_VERSION: "1.21"
  BUILD_ENV: "production"
```

Docker build arguments (passed with `docker build --build-arg`). Useful for parameterizing Dockerfiles.

#### `target`

```yaml
target: production
```

For multi-stage Dockerfiles, specifies which stage to build. Equivalent to `docker build --target=production`.

#### `rebuildStrategy`

```yaml
rebuildStrategy: default  # default | always | ignoreContextChanges
```

Controls when DevSpace rebuilds the image:
- **`default`**: Rebuild if Dockerfile or context files changed.
- **`always`**: Always rebuild (skip caching).
- **`ignoreContextChanges`**: Only rebuild if Dockerfile changed (ignore source code changes).

## When Are Images Built?

Images are built when you run a pipeline that calls `build_images`.

### During `devspace deploy`

The `deploy` pipeline includes:

```yaml
build_images --all -t $(git describe --always)
```

This builds all images in the `images` section and tags them with the git commit hash.

### During `devspace dev`

The `dev` pipeline **does not** call `build_images` by default:

```yaml
run_dependencies --all
create_deployments --all   # No build_images!
start_dev app
```

**Why?** In dev mode, DevSpace replaces your production image with a `devImage` (from the `dev` section). This skips the build to speed up the dev loop.

If you want to build images in dev mode, you can customize the `dev` pipeline:

```yaml
pipelines:
  dev:
    run: |-
      build_images --all
      create_deployments --all
      start_dev app
```

But this slows down `devspace dev`. The recommended approach: use `devImage` and file sync (as we did in Chapter 4).

### Manual Build

You can build images without deploying:

```bash
devspace build
```

This runs the `build` pipeline (default: `build_images --all`).

## Image Tagging

DevSpace automatically tags images. Let's see how.

### Default Tag

By default, DevSpace tags images with a **random hash** based on the build context and Dockerfile:

```
ghcr.io/loft-sh/devspace-quickstart-golang:abc1234
```

The hash changes only if the Dockerfile or context files change. This acts as a cache key: if nothing changed, DevSpace skips the build.

### Custom Tags

You can override tags with the `-t` flag in the pipeline:

```yaml
build_images --all -t $(git describe --always)
```

This tags the image with the git commit hash (e.g., `v1.2.3-5-gabc1234`).

**Other examples:**

- **Semantic version**: `-t v1.0.0`
- **Branch name**: `-t $(git rev-parse --abbrev-ref HEAD)`
- **Timestamp**: `-t $(date +%Y%m%d-%H%M%S)`

### Where Tags Are Used

When you deploy, DevSpace injects the tag into the deployment values.

**Example:**

1. You build `ghcr.io/.../app` and DevSpace tags it `abc1234`.
2. In `devspace.yaml`, the deployment has:

```yaml
deployments:
  app:
    helm:
      values:
        containers:
        - image: ghcr.io/loft-sh/devspace-quickstart-golang
```

3. DevSpace replaces this with:

```yaml
containers:
- image: ghcr.io/loft-sh/devspace-quickstart-golang:abc1234
```

4. Helm deploys with the correct tag.

This ensures your deployment always uses the image you just built.

## Hands-On: Add a Build Argument

Let's customize the Dockerfile with a build argument.

### Step 1: Add a Build Arg to the Dockerfile

Open `Dockerfile` and add a build argument for the Go version:

```dockerfile
ARG GO_VERSION=1.21

FROM golang:${GO_VERSION}-alpine AS builder
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

Now you can control the Go version at build time.

### Step 2: Add buildArgs to devspace.yaml

Open `devspace.yaml` and add `buildArgs` to the `images.app` section:

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    dockerfile: ./Dockerfile
    buildArgs:
      GO_VERSION: "1.22"
```

### Step 3: Rebuild

```bash
devspace build
```

DevSpace builds the image with `--build-arg GO_VERSION=1.22`. You'll see in the output:

```
build:image app Building image 'ghcr.io/.../app:xyz' with engine 'docker'
build:image app [+] Building 12.3s (14/14) FINISHED
...
```

The image now uses Go 1.22.

### Step 4: Verify (Optional)

Deploy and check the Go version in the container:

```bash
devspace deploy
devspace enter
```

In the container:

```bash
go version
```

You should see Go 1.22.

## Build Engines

DevSpace supports multiple build engines:

1. **Docker** (default): Uses your local Docker daemon. Fast for local dev.
2. **Kaniko**: Builds images inside Kubernetes (no local Docker needed). Useful for CI or rootless environments.
3. **Custom**: Run any command (e.g., Buildah, Cloud Build).

### Using Kaniko

To build with Kaniko, add `buildKit` or `kaniko` config:

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    dockerfile: ./Dockerfile
    kaniko:
      cache: true
```

DevSpace will build the image in a Kubernetes pod using Kaniko. Slower but works anywhere.

### Custom Build Command

For full control:

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    custom:
      command: ./build.sh
      args: ["${runtime.images.app.image}", "${runtime.images.app.tag}"]
```

DevSpace runs your custom script to build the image.

## Registries and Push Behavior

### Local Clusters (No Push)

For **minikube** or **kind**, you typically don't push images to a registry. DevSpace builds images and makes them available to the cluster directly.

**minikube:**
```bash
eval $(minikube docker-env)
devspace deploy
```

This builds images in minikube's Docker daemon. No push needed.

**kind:**
```bash
devspace deploy
kind load docker-image ghcr.io/loft-sh/devspace-quickstart-golang:abc1234 --name devspace-tutorial
```

DevSpace can auto-detect kind and load images for you.

### Remote Clusters (Push Required)

For **cloud clusters** (GKE, EKS, AKS), you must push images to a registry the cluster can access.

**Steps:**

1. **Choose a registry** (Docker Hub, GCR, ECR, ACR, GitHub Container Registry, etc.).
2. **Authenticate Docker** to the registry:

```bash
docker login ghcr.io  # For GitHub Container Registry
# Or: gcloud auth configure-docker  # For GCR
# Or: aws ecr get-login-password | docker login --username AWS --password-stdin <ecr-url>
```

3. **Update `devspace.yaml`** with your registry:

```yaml
images:
  app:
    image: ghcr.io/myusername/myapp  # Your registry
    dockerfile: ./Dockerfile
```

4. **Deploy:**

```bash
devspace deploy
```

DevSpace builds, tags, pushes the image, then deploys.

### Pull Secrets

If your registry requires authentication (private images), Kubernetes needs a pull secret.

DevSpace can create pull secrets automatically:

```bash
devspace deploy
```

If DevSpace detects a private registry, it prompts:

```
? Do you want to create a pull secret for this registry?
  > Yes
    No
```

Say **Yes**. DevSpace creates a Kubernetes secret with your Docker credentials and adds it to the deployment.

Alternatively, create the secret manually:

```bash
kubectl create secret docker-registry my-registry-secret \
  --docker-server=ghcr.io \
  --docker-username=myusername \
  --docker-password=mytoken \
  -n devspace-tutorial
```

Then reference it in `devspace.yaml`:

```yaml
deployments:
  app:
    helm:
      values:
        pullSecrets:
        - my-registry-secret
```

## Hands-On: Add a Second Image

Let's add a second image (e.g., a sidecar or helper) and see parallel building.

### Step 1: Create a Simple Dockerfile

Create a new file `Dockerfile.sidecar`:

```dockerfile
FROM alpine:latest
RUN echo "Hello from sidecar" > /hello.txt
CMD ["cat", "/hello.txt"]
```

This is a minimal image that just prints a message.

### Step 2: Add the Image to devspace.yaml

```yaml
images:
  app:
    image: ghcr.io/loft-sh/devspace-quickstart-golang
    dockerfile: ./Dockerfile
  sidecar:
    image: ghcr.io/loft-sh/devspace-quickstart-golang-sidecar
    dockerfile: ./Dockerfile.sidecar
```

### Step 3: Build Both Images

```bash
devspace build
```

You'll see DevSpace build both images **in parallel**:

```
build:image app Building image 'ghcr.io/.../app:xyz'
build:image sidecar Building image 'ghcr.io/.../sidecar:abc'
...
```

Parallel builds speed things up when you have multiple images.

### Step 4: Add the Sidecar to the Deployment (Optional)

If you want to deploy the sidecar alongside the main app:

```yaml
deployments:
  app:
    helm:
      values:
        containers:
        - image: ghcr.io/loft-sh/devspace-quickstart-golang
        - image: ghcr.io/loft-sh/devspace-quickstart-golang-sidecar
          name: sidecar
```

Run `devspace deploy`. Both containers run in the same pod.

## What You Learned

- The `images` section defines which images to build, with options for `dockerfile`, `context`, `buildArgs`, `target`, and `rebuildStrategy`
- Images are built when a pipeline calls `build_images` (e.g., in the `deploy` pipeline)
- DevSpace automatically tags images (default: hash of context); you can customize with `-t` in pipelines
- Tags are injected into deployments so Helm/kubectl uses the correct image
- For local clusters (minikube/kind), you typically don't push images; for remote clusters, you push to a registry and may need pull secrets
- DevSpace builds multiple images in parallel for speed

## Next Steps

You now know how to build and customize images. Let's make the config more flexible with variables and profiles. Move on to [Chapter 6: Variables and Profiles](../06-variables-profiles/README.md).

## Troubleshooting

**Problem: Build fails with "dockerfile not found"**
- Check the `dockerfile` path is correct and relative to the project root
- Make sure the file exists: `ls -la Dockerfile`

**Problem: Image pull error (ErrImagePull, ImagePullBackOff)**
- For local clusters: Make sure you're not pushing to a registry (use "Skip Registry" in `devspace init`)
- For remote clusters: Make sure the image was pushed and the cluster can access the registry
- Check pull secrets if using a private registry

**Problem: Build is slow**
- Use Docker BuildKit for faster builds: `export DOCKER_BUILDKIT=1`
- Add a `.dockerignore` file to exclude unnecessary files from the build context
- Use multi-stage Dockerfiles to minimize final image size

**Problem: Tag doesn't change even though I edited code**
- DevSpace caches builds. To force rebuild: `devspace build --force-rebuild`
- Or: `devspace purge && devspace deploy`
