# Chapter 4: Development Mode (devspace dev)

## Learning Objectives

By the end of this chapter, you will:
- Understand what `devspace dev` does and how it differs from `devspace deploy`
- Use file synchronization to see code changes instantly without rebuilding images
- Use port forwarding to access your app on localhost
- Use the DevSpace UI to view logs, open terminals, and manage pods
- Know when to use `devspace logs` and `devspace enter`

## Prerequisites

- Completed [Chapter 3: Understanding devspace.yaml](../03-understanding-config/README.md)
- Have the `devspace-quickstart-golang` project initialized and deployed

## What is `devspace dev`?

`devspace dev` is DevSpace's **development mode**. It's designed for the inner loop of coding:

1. **Deploy your app** to Kubernetes
2. **Start file sync** (local files ↔ container files)
3. **Forward ports** (access the app on localhost)
4. **Open a terminal** to the container
5. **Keep watching** for changes (and sync them automatically)

The key difference from `devspace deploy`:
- **`devspace deploy`**: Builds production images, deploys them, then exits.
- **`devspace dev`**: Skips the image build (uses a `devImage` with dev tools), deploys, and **stays running** to sync files and forward ports.

This gives you **hot reloading**: change code locally, see it update in Kubernetes instantly.

## Start Dev Mode

From the `devspace-quickstart-golang` directory:

```bash
devspace dev
```

**What happens:**

1. **DevSpace runs the `dev` pipeline**:
   - Deploy dependencies (if any)
   - Create deployments (replaces your image with `devImage`)
   - Start dev mode for `app`

2. **DevSpace starts file sync**: Watches your local files and syncs changes to the container.

3. **DevSpace forwards ports**: `2345` (debugger) and `8080` (HTTP) are forwarded to localhost.

4. **DevSpace opens a terminal** to the container and runs `./devspace_start.sh`.

You'll see output like:

```
info Using namespace 'devspace-tutorial'
info Using kube context 'minikube'
deploy:app Deploying with helm...
dev:app Waiting for pod to become ready...
dev:app Port forwarding started on 2345:2345
dev:app Port forwarding started on 8080:8080
dev:app Sync started on ./
dev:app Terminal opened

    ____                 _____
   / __ \               / ____|
  | |  | |_ __   ___ __| (___  _ __   __ _  ___ ___
  | |  | | '_ \ / _ \ '_ \___| | '_ \ / _` |/ __/ _ \
  | |__| | |_) |  __/ | | |____| |_) | (_| | (_|  __/
   \____/| .__/ \___|_| |_|_____| .__/ \__,_|\___\___|
         | |                    | |
         |_|                    |_|

Welcome to your development container!

This is how you can work with it:
- Run `go run main.go` to start the application
- Files will be synchronized between your local machine and this container
- Some ports will be forwarded, so you can access the application on localhost

root@app-xyz:/app#
```

You're now **inside the container** with an interactive terminal!

## Start the Application

The container has your source code (synced from your local machine). Let's start the app:

```bash
go run main.go
```

You'll see:

```
Server listening on :8080
```

The app is now running inside the container. Because port `8080` is forwarded, you can access it on your laptop.

**Open your browser** and visit:

```
http://localhost:8080
```

You should see:

```
Hello from DevSpace!
```

Great! The app is running in Kubernetes, but you're accessing it as if it were running locally.

## Hot Reloading with File Sync

Now for the magic: let's change the code without rebuilding the image.

### Edit the Code

**Keep `devspace dev` running** (and `go run main.go` running in the container).

In a **new terminal** on your local machine, open `main.go`:

```bash
cd devspace-quickstart-golang
nano main.go  # or use your favorite editor
```

Change the message:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from DevSpace! (Hot reloaded!)")  // Changed
}
```

**Save the file.**

### Watch the Sync

In the `devspace dev` terminal, you'll see:

```
dev:app Sync: ↑ main.go
```

DevSpace detected the change and uploaded `main.go` to the container.

### Restart the App

In the container terminal (where `go run main.go` is running), press `Ctrl+C` to stop the app, then restart it:

```bash
go run main.go
```

The app recompiles (inside the container, using the `devImage`'s Go toolchain) and starts.

**Refresh your browser** (`http://localhost:8080`). You'll see:

```
Hello from DevSpace! (Hot reloaded!)
```

**You just deployed a code change to Kubernetes without rebuilding the Docker image or redeploying.** This is the power of DevSpace's dev mode.

### Automatic Restart (Optional)

For Go, you can use a file watcher like [air](https://github.com/cosmtrek/air) to automatically restart the app when files change. For Node.js, `nodemon` does this. For Python, `watchdog` or `flask --reload`. DevSpace syncs the files; the file watcher restarts the app.

To try `air` in this project:

1. Install `air` in the container:

```bash
go install github.com/cosmtrek/air@latest
```

2. Run:

```bash
~/go/bin/air
```

Now when you edit `main.go` locally, `air` will detect the sync and restart the app automatically. (This is beyond the scope of this chapter, but good to know!)

## Port Forwarding

Port forwarding is how you access services running in Kubernetes from your local machine.

DevSpace automatically forwards the ports defined in `dev.app.ports`:

```yaml
ports:
- port: "2345"   # Debugger port
- port: "8080"   # HTTP port
```

### HTTP Port (8080)

We've already used this: `http://localhost:8080` accesses the app.

### Debugger Port (2345)

If you're using a Go debugger (like Delve), you can attach to port `2345`. This is useful for IDE debugging (e.g., VS Code's "Attach to Process" feature).

### Custom Port Mapping

You can map a different local port to the container port:

```yaml
ports:
- port: "3000:8080"  # localhost:3000 → container:8080
```

Now you'd visit `http://localhost:3000`.

### Reverse Port Forwarding

DevSpace also supports **reverse port forwarding** (container → localhost). This is useful if your container needs to access a service running on your laptop (e.g., a local database).

```yaml
reversePorts:
- port: "5432:5432"  # container:5432 → localhost:5432
```

We won't use this in this tutorial, but it's available.

## The DevSpace UI

When `devspace dev` is running, DevSpace starts a **localhost web UI** (default port 8090). This UI lets you view logs, open terminals, and manage pods from your browser.

### Open the UI

While `devspace dev` is running, open:

```
http://localhost:8090
```

(If port 8090 is taken, DevSpace uses a different port. Check the `devspace dev` output for the actual URL.)

### UI Features

The DevSpace UI shows:

1. **Pods**: List of all pods in your namespace, with status indicators (Running, Pending, Failed).
2. **Logs**: Click a pod to stream logs in real time.
3. **Terminal**: Click "Terminal" to open a web-based terminal to any container.
4. **Port Forwarding**: Click "Open" next to a port-forwarded URL (e.g., `http://localhost:8080`) to open it in a new tab.

**Try it:**
- Click on the `app` pod.
- You'll see logs from the app (the `go run main.go` output).
- Click "Terminal" to open a new terminal to the container (in your browser, no separate terminal window needed).

The UI is especially useful when working with multiple pods or services.

## Stopping Dev Mode

To stop `devspace dev`, press `Ctrl+C` in the terminal where it's running.

DevSpace will:
1. Stop file sync
2. Stop port forwarding
3. Close the terminal
4. (Optional) Stop or remove the dev container (depending on config)

The app stays deployed in the cluster. If you want to remove it, run:

```bash
devspace purge
```

## Other Useful Commands

Sometimes you don't want full dev mode (terminal, sync, etc.). DevSpace has individual commands for common tasks.

### `devspace logs`

Stream logs from your app without starting dev mode:

```bash
devspace logs
```

DevSpace shows a list of pods. Select one, and it streams logs. Press `Ctrl+C` to stop.

**Useful when:**
- You just want to check logs quickly.
- The app is already deployed (via `devspace deploy`) and you don't need sync/ports.

### `devspace enter`

Open a terminal to a container without starting dev mode:

```bash
devspace enter
```

DevSpace shows a list of pods. Select one, and it opens a shell. Type `exit` to close.

**Useful when:**
- You want to run a one-off command (e.g., database migration, debugging).
- You don't need file sync.

### `devspace sync`

Start only file synchronization (no terminal, no port forwarding):

```bash
devspace sync
```

**Useful when:**
- You want sync but not the other dev features.
- You're using a separate tool for port forwarding or terminal.

### `devspace ui`

Open just the DevSpace UI (without starting dev mode):

```bash
devspace ui
```

The UI opens and connects to your cluster. You can view pods, logs, and terminals.

## Hands-On: Customize Sync

Let's exclude some files from syncing. You probably don't want to sync large build artifacts or `node_modules` (for Node projects).

### Edit devspace.yaml

Open `devspace.yaml` and find the `dev.app.sync` section:

```yaml
sync:
- path: ./
```

Add an `excludePaths` list:

```yaml
sync:
- path: ./
  excludePaths:
  - .git/
  - .devspace/
  - '*.log'
```

This tells DevSpace **not to sync**:
- The `.git/` directory
- The `.devspace/` directory
- Any `.log` files

### Restart Dev Mode

Stop `devspace dev` (`Ctrl+C`) and restart:

```bash
devspace dev
```

Now if you create a `test.log` file locally, DevSpace won't sync it. This speeds up sync and avoids clutter.

## What You Learned

- `devspace dev` deploys your app with a `devImage` (includes dev tools) and keeps running to sync files and forward ports
- **File sync** (local ↔ container) enables hot reloading without rebuilding images
- **Port forwarding** lets you access apps in Kubernetes on localhost
- The **DevSpace UI** (port 8090 by default) provides a web-based view of pods, logs, and terminals
- `devspace logs`, `devspace enter`, and `devspace sync` are shortcuts for specific dev tasks

## Next Steps

Now that you know how to develop efficiently, let's dive deeper into image building. Move on to [Chapter 5: Images and Builds](../05-images-builds/README.md) to learn about build arguments, tagging, and when images are built.

## Troubleshooting

**Problem: Port 8080 already in use**
- Stop the other process: `lsof -i :8080` (macOS/Linux) or `netstat -ano | findstr :8080` (Windows)
- Or change the port in `devspace.yaml` (e.g., use `3000:8080` in `dev.app.ports`)

**Problem: File sync is slow**
- Exclude large directories (e.g., `node_modules/`, `vendor/`, `.git/`) in `dev.app.sync.excludePaths`
- Check your network latency if using a remote cluster

**Problem: Changes don't appear in the container**
- Make sure file sync is running (check `devspace dev` output for "Sync started")
- Check if the file is excluded (see `excludePaths`)
- Restart dev mode: `Ctrl+C` then `devspace dev` again

**Problem: Can't access the UI (localhost:8090)**
- Check the `devspace dev` output for the actual UI port (might be different if 8090 is taken)
- Make sure `devspace dev` is still running
