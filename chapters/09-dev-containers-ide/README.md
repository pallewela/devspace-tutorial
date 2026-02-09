# Chapter 9: Dev Containers and IDE Integration

## Learning Objectives

By the end of this chapter, you will:
- Understand all options in the `dev` section (sync, ports, terminal, SSH, etc.)
- Configure multiple sync paths and exclusions
- Use SSH to connect your IDE (VS Code, JetBrains) to dev containers
- Define custom commands for team-shared workflows
- Know when to use `terminal` vs `logs` mode

## Prerequisites

- Completed [Chapter 8: Dependencies](../08-dependencies/README.md)
- Have the `devspace-quickstart-golang` project initialized

## The `dev` Section

The `dev` section defines how DevSpace behaves during `devspace dev`. Let's explore all available options.

### Basic Structure

```yaml
dev:
  <name>:
    imageSelector: <image>
    devImage: <dev-image>
    ports: [...]
    sync: [...]
    terminal: {...}
    ssh: {...}
    open: [...]
    proxyCommands: [...]
```

Each dev configuration has a **name** (e.g., `app`, `api`, `frontend`). You can have multiple dev configs and start them individually or all together.

## Container Selection

DevSpace needs to know which container to target.

### `imageSelector`

```yaml
imageSelector: ghcr.io/loft-sh/devspace-quickstart-golang
```

Select the container(s) running this image. DevSpace finds all pods with a container matching this image.

### `labelSelector`

```yaml
labelSelector: app=myapp,tier=frontend
```

Select containers by Kubernetes labels. Useful when multiple images share the same labels.

### `container`

```yaml
container: mycontainer
```

If a pod has multiple containers, specify which one to target.

## Dev Image Replacement

### `devImage`

```yaml
devImage: ghcr.io/loft-sh/devspace-containers/go:1.21-alpine
```

Replace the production image with a dev image. The dev image should include:
- **Development tools** (compilers, debuggers, hot-reload tools)
- **Shell** (bash, sh) for terminal access
- **Source code tools** (git, package managers)

**Why replace?**
- Your production image is minimal (just the compiled binary, no Go toolchain).
- The dev image includes everything you need to build and run the app inside the container.
- You skip rebuilding the image every time you change code.

**How it works:**
1. DevSpace deploys your app.
2. DevSpace replaces the container image with `devImage`.
3. DevSpace syncs your source code to the container.
4. You build and run the app inside the container (using the dev image's tools).

### `replaceImage`

```yaml
replaceImage: my-registry/myapp
```

Explicitly specify which image to replace (if `imageSelector` matches multiple containers).

## File Synchronization

The `sync` section defines which files to sync between your local machine and the container.

### Basic Sync

```yaml
sync:
- path: ./
```

Sync everything in the project root (`.`) to the container's working directory.

### Path Mapping

```yaml
sync:
- path: ./src:/app/src
```

Sync local `./src` to container `/app/src`.

### Multiple Sync Entries

```yaml
sync:
- path: ./src:/app/src
- path: ./config:/app/config
- path: ./public:/app/public
```

Sync multiple directories.

### Exclude Paths

```yaml
sync:
- path: ./
  excludePaths:
  - .git/
  - node_modules/
  - .devspace/
  - '*.log'
  - tmp/
```

Don't sync these files or directories. This speeds up sync and avoids clutter.

### Upload-Only and Download-Only Exclusions

```yaml
sync:
- path: ./
  uploadExcludePaths:
  - dist/    # Don't upload build artifacts (generate them in container)
  downloadExcludePaths:
  - .env     # Don't download environment files from container
```

- **`uploadExcludePaths`**: Don't upload these files from local → container.
- **`downloadExcludePaths`**: Don't download these files from container → local.

### Disable File Watching

```yaml
sync:
- path: ./config:/app/config
  noWatch: true
```

Sync once but don't watch for changes. Useful for config files that rarely change.

### One-Way Sync

```yaml
sync:
- path: ./logs:/app/logs
  disableUpload: true
```

Only sync from container → local (download only). Useful for log files.

```yaml
sync:
- path: ./src:/app/src
  disableDownload: true
```

Only sync from local → container (upload only). Most common use case.

## Port Forwarding

### Forward Ports

```yaml
ports:
- port: "8080"          # localhost:8080 → container:8080
- port: "3000:8080"     # localhost:3000 → container:8080
```

### Reverse Port Forwarding

```yaml
reversePorts:
- port: "5432:5432"     # container:5432 → localhost:5432
```

The container can access services running on your local machine. Useful if:
- Your local machine has a database the container needs to connect to
- You're debugging with a remote debugger on your local machine

## Terminal

### Interactive Terminal

```yaml
terminal:
  enabled: true
  command: ./devspace_start.sh
```

DevSpace opens a terminal to the container and runs the command. This is the default behavior.

### Disable Terminal (Logs Mode)

```yaml
terminal:
  enabled: false
```

Instead of opening a terminal, DevSpace streams logs from the container. Useful if:
- You don't need an interactive session
- You're running in CI
- You prefer to use `devspace enter` for manual terminal access

### Custom Working Directory

```yaml
terminal:
  workDir: /app/backend
```

Start the terminal in a specific directory.

## SSH

DevSpace can inject a lightweight SSH server into the container. Your IDE (VS Code, JetBrains, etc.) can connect to it for "remote development."

### Enable SSH

```yaml
ssh:
  enabled: true
```

DevSpace injects an SSH server and forwards a local port (e.g., `localhost:2222`) to it.

### Use with VS Code

**Install the Remote-SSH extension in VS Code:**
1. Open VS Code
2. Install "Remote - SSH" extension

**Connect to the dev container:**
1. Run `devspace dev`
2. In VS Code, open the Command Palette (`Cmd+Shift+P` or `Ctrl+Shift+P`)
3. Select "Remote-SSH: Connect to Host"
4. Enter `devspace.<namespace>.<pod-name>` (DevSpace provides the exact command in the output)

VS Code connects to the container. You can edit files, run terminals, and debug—all inside the container.

### Proxy Commands

```yaml
proxyCommands:
- command: devspace
- command: kubectl
- command: helm
- command: git
```

When you run these commands inside the container, they're proxied to your **local machine**. So `kubectl get pods` inside the container runs on your laptop (using your local kubectl config).

**Why?**
- The container doesn't need kubectl installed.
- You use your local credentials and context.
- You can run `devspace` commands from inside the container.

## Auto-Open URLs

```yaml
open:
- url: http://localhost:8080
- url: http://localhost:3000
  cmd: echo "Frontend ready!"
```

DevSpace automatically opens these URLs in your browser once the container is ready. The `cmd` field is optional (run a command instead of opening a browser).

## Hands-On: Advanced Sync Configuration

Let's configure sync for a more complex project.

### Scenario

You have:
- Source code in `./src`
- Config files in `./config`
- Build artifacts in `./dist` (generated by the container, don't sync from local)
- Logs in `./logs` (generated by the container, download but don't upload)

### Configuration

Open `devspace.yaml` and update the `dev.app.sync` section:

```yaml
dev:
  app:
    sync:
    - path: ./src:/app/src
      excludePaths:
      - '*.tmp'
      - .DS_Store
    - path: ./config:/app/config
      noWatch: true          # Config rarely changes, sync once
    - path: ./dist:/app/dist
      disableUpload: true    # Only download build artifacts
    - path: ./logs:/app/logs
      disableUpload: true    # Only download logs
```

### Test It

1. Run `devspace dev`
2. Create a file locally: `echo "test" > src/test.txt`
   - DevSpace syncs it to the container
3. In the container, create a build artifact: `echo "build" > dist/output.js`
   - DevSpace syncs it to your local `dist/` folder
4. In the container, create a log: `echo "error" > logs/app.log`
   - DevSpace syncs it to your local `logs/` folder

## Hands-On: Custom Commands

Custom commands let you define team-shared workflows (e.g., run migrations, seed database, run tests).

### Define Commands

Add a `commands` section to `devspace.yaml`:

```yaml
commands:
  migrate:
    command: |-
      echo "Running database migrations..."
      devspace enter -- go run migrations/migrate.go
  
  test:
    command: |-
      echo "Running tests in container..."
      devspace enter -- go test ./...
  
  seed:
    command: |-
      echo "Seeding database..."
      devspace enter -- go run scripts/seed.go
```

### Run Commands

```bash
devspace run migrate
devspace run test
devspace run seed
```

Each command runs in the container (via `devspace enter`). Your team can use the same commands without remembering complex kubectl or go commands.

### Commands with Flags

```yaml
commands:
  test:
    flags:
    - name: verbose
      short: v
      type: bool
      description: "Verbose test output"
    command: |-
      if [ "$(devspace get flag verbose)" = "true" ]; then
        devspace enter -- go test -v ./...
      else
        devspace enter -- go test ./...
      fi
```

Run:

```bash
devspace run test --verbose
```

## Multiple Dev Configurations

You can define multiple dev configs and start them individually.

### Example: API and Frontend

```yaml
dev:
  api:
    imageSelector: my-registry/api
    ports:
    - port: "8080"
    sync:
    - path: ./api:/app
  
  frontend:
    imageSelector: my-registry/frontend
    ports:
    - port: "3000"
    sync:
    - path: ./frontend:/app
```

### Start Both

```bash
devspace dev
```

DevSpace starts both dev configurations (both terminals, both syncs, both port-forwards).

### Start Only One

```bash
devspace dev api
```

Starts only the API dev config.

## What You Learned

- The `dev` section has many options: `imageSelector`, `devImage`, `sync`, `ports`, `terminal`, `ssh`, `open`, `proxyCommands`
- **File sync** supports multiple paths, exclusions, one-way sync, and upload/download filters
- **SSH** lets your IDE connect to the dev container for remote development
- **Proxy commands** let you run local tools (kubectl, devspace, git) from inside the container
- **Custom commands** provide team-shared workflows (migrate, test, seed, etc.)
- You can define **multiple dev configs** and start them individually or together

## Next Steps

You're now an expert at dev containers and IDE integration. Let's see how to use DevSpace in CI/CD and how to clean up resources. Move on to [Chapter 10: CI/CD and Cleanup](../10-cicd-cleanup/README.md).

## Troubleshooting

**Problem: Sync is slow**
- Add exclusions for large directories (node_modules, vendor, .git, dist, logs)
- Use `noWatch: true` for rarely-changing files
- Check network latency (especially for remote clusters)

**Problem: Files not syncing**
- Check if the file is excluded (see `excludePaths`)
- Restart dev mode: `Ctrl+C`, then `devspace dev` again
- Check container file system: `devspace enter` then `ls -la`

**Problem: SSH not working with VS Code**
- Make sure `ssh.enabled: true` in `devspace.yaml`
- Check the DevSpace output for the exact SSH command to use
- Restart dev mode and try again

**Problem: Terminal doesn't open**
- Check if `terminal.enabled: true` (or not explicitly set to `false`)
- Check if the container's entrypoint is blocking (try `terminal.command: /bin/sh`)
- Check container logs: `devspace logs`
