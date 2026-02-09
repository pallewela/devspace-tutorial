# Chapter 7: Pipelines

## Learning Objectives

By the end of this chapter, you will:
- Understand what pipelines are and how they work
- Know the built-in pipeline functions (`build_images`, `create_deployments`, `start_dev`, etc.)
- Customize existing pipelines (dev, deploy, build, purge)
- Create custom pipelines and run them with `devspace run-pipeline`
- Add custom flags to pipelines and use them in scripts
- Use global helper functions (`is_equal`, `is_empty`, `get_flag`)

## Prerequisites

- Completed [Chapter 6: Variables and Profiles](../06-variables-profiles/README.md)
- Have the `devspace-quickstart-golang` project initialized

## What Are Pipelines?

Pipelines are **scripts** in `devspace.yaml` that define what happens when you run DevSpace commands. They're written in **POSIX shell syntax** but run cross-platform (DevSpace emulates bash on Windows/macOS/Linux).

Pipelines use special **built-in functions** like `build_images`, `create_deployments`, and `start_dev` to tell DevSpace when to build, deploy, or start dev mode.

### Why Pipelines?

Before pipelines (DevSpace v5 and earlier), DevSpace had a fixed workflow: build → deploy → dev. You couldn't customize the order or add custom steps.

With pipelines (DevSpace v6+), you have full control:
- Build only some images
- Deploy in a specific order
- Run tests between build and deploy
- Skip steps based on conditions
- Add custom commands

## Default Pipelines

DevSpace defines four default pipelines that map to commands:

| Pipeline | Command | Default Workflow |
|----------|---------|------------------|
| `dev` | `devspace dev` | `run_dependencies → create_deployments → start_dev` |
| `deploy` | `devspace deploy` | `run_dependencies → build_images → create_deployments` |
| `build` | `devspace build` | `build_images` |
| `purge` | `devspace purge` | `stop_dev → purge_deployments → run_dependencies --pipeline purge` |

When you run a command like `devspace deploy`, DevSpace executes the `deploy` pipeline.

### The `deploy` Pipeline

```yaml
pipelines:
  deploy:
    run: |-
      run_dependencies --all
      build_images --all -t $(git describe --always)
      create_deployments --all
```

**What happens:**

1. **`run_dependencies --all`**: If your project has dependencies (other services), deploy them first.
2. **`build_images --all -t $(git describe --always)`**: Build all images and tag with git commit hash.
3. **`create_deployments --all`**: Deploy all entries in `deployments`.

Simple and sequential.

### The `dev` Pipeline

```yaml
pipelines:
  dev:
    run: |-
      run_dependencies --all
      create_deployments --all
      start_dev app
```

**Difference from `deploy`:**
- **No `build_images`**: Dev mode uses `devImage`, so building is skipped.
- **`start_dev app`**: Start dev features (sync, ports, terminal) for the `app` dev config.

## Built-In Pipeline Functions

DevSpace provides functions you can call in pipelines. Let's explore the most important ones.

### Image Functions

#### `build_images [image-1] [image-2] ... [--flags]`

Builds specified images. Use `--all` to build all images.

**Flags:**
- `--all`: Build all images in the `images` section
- `-t TAG` / `--tag TAG`: Override the image tag
- `--skip`: Skip building (useful in conditionals)
- `--force-rebuild`: Force rebuild even if nothing changed
- `--sequential`: Build images one at a time (default: parallel)

**Examples:**

```bash
build_images --all -t v1.0.0
build_images app api  # Build only 'app' and 'api' images
```

### Deployment Functions

#### `create_deployments [deployment-1] [deployment-2] ... [--flags]`

Deploys specified deployments. Use `--all` to deploy all.

**Flags:**
- `--all`: Deploy all deployments
- `--force-redeploy`: Force redeploy even if nothing changed
- `--sequential`: Deploy one at a time (default: parallel)

**Examples:**

```bash
create_deployments --all
create_deployments app database  # Deploy only 'app' and 'database'
```

#### `purge_deployments [deployment-1] [deployment-2] ... [--flags]`

Deletes specified deployments.

**Flags:**
- `--all`: Purge all deployments
- `--force-purge`: Purge even if other projects might depend on them

**Example:**

```bash
purge_deployments --all
```

### Dev Functions

#### `start_dev [dev-1] [dev-2] ... [--flags]`

Starts dev mode for specified dev configs.

**Flags:**
- `--all`: Start all dev configs
- `--disable-sync`: Skip file sync
- `--disable-port-forwarding`: Skip port forwarding
- `--disable-open`: Don't auto-open URLs

**Examples:**

```bash
start_dev app
start_dev --all --disable-sync  # Start dev without file sync
```

#### `stop_dev [dev-1] [dev-2] ... [--flags]`

Stops dev mode.

**Flags:**
- `--all`: Stop all dev configs

### Pipeline Functions

#### `run_pipelines [pipeline-1] [pipeline-2] ... [--flags]`

Runs other pipelines.

**Flags:**
- `--background`: Run in the background (parallel)
- `--sequential`: Run one after another

**Examples:**

```bash
run_pipelines build test  # Run 'build' pipeline, then 'test' pipeline
run_pipelines test-1 test-2 --background  # Run both tests in parallel
```

#### `run_default_pipeline [pipeline]`

Runs the default (built-in) version of a pipeline, even if you've overridden it.

**Example:**

```yaml
pipelines:
  deploy:
    run: |-
      echo "Running custom pre-deploy checks..."
      run_default_pipeline deploy  # Now run the default deploy logic
```

### Dependency Functions

#### `run_dependencies [dep-1] [dep-2] ... [--flags]`

Deploys specified dependencies.

**Flags:**
- `--all`: Deploy all dependencies
- `--pipeline NAME`: Which pipeline to run for each dependency (default: `deploy`)

**Example:**

```bash
run_dependencies --all --pipeline dev  # Run the 'dev' pipeline for all dependencies
```

### Utility Functions

#### `exec_container [command] [--flags]`

Executes a command inside a container.

**Flags:**
- `--image-selector IMAGE`: Select container by image name
- `--label-selector LABEL`: Select container by label
- `--container NAME`: Select specific container in a pod

**Example:**

```bash
exec_container --image-selector ghcr.io/.../app -- ls -la
exec_container --image-selector ghcr.io/.../app -- go test ./...
```

#### `wait_pod [--flags]`

Waits for a pod to be ready.

**Flags:**
- `--image-selector IMAGE`
- `--timeout SECONDS`

**Example:**

```bash
wait_pod --image-selector ghcr.io/.../app --timeout 120
```

## Global Helper Functions

These functions can be used anywhere in `devspace.yaml` (not just in pipelines).

### `is_equal [value-1] [value-2]`

Returns exit code 0 (success) if values are equal.

**Example:**

```bash
if is_equal ${DEVSPACE_NAMESPACE} "production"; then
  echo "Deploying to production!"
fi
```

### `is_empty [value]`

Returns exit code 0 if value is empty.

**Example:**

```bash
if is_empty ${MY_VAR}; then
  echo "MY_VAR is not set"
fi
```

### `get_flag [flag-name]`

Gets the value of a custom flag (more below).

**Example:**

```bash
TAG=$(get_flag "tag")
if ! is_empty $TAG; then
  build_images --all -t $TAG
fi
```

### `is_true [value]`

Returns exit code 0 if value is `"true"`.

**Example:**

```bash
if is_true $(get_flag "verbose"); then
  echo "Verbose mode enabled"
fi
```

## Hands-On: Customize the Deploy Pipeline

Let's add a step to the `deploy` pipeline that runs tests before deploying.

### Step 1: Add a Test Command

For simplicity, we'll use `go test`. Open `devspace.yaml` and update the `deploy` pipeline:

```yaml
pipelines:
  deploy:
    run: |-
      echo "Running tests..."
      go test ./...
      
      if [ $? -ne 0 ]; then
        echo "Tests failed! Aborting deploy."
        exit 1
      fi
      
      echo "Tests passed. Deploying..."
      run_dependencies --all
      build_images --all -t $(git describe --always)
      create_deployments --all
```

### Step 2: Deploy

```bash
devspace deploy
```

DevSpace runs `go test ./...`. If tests fail (exit code ≠ 0), the pipeline stops. If tests pass, deployment continues.

**Note:** This runs tests on your local machine. For in-cluster tests, use `exec_container` (see below).

## Hands-On: Create a Custom Pipeline

Let's create a pipeline called `test` that deploys the app and runs tests inside the container.

### Step 1: Define the Pipeline

Add a new pipeline to `devspace.yaml`:

```yaml
pipelines:
  test:
    run: |-
      echo "Deploying app for testing..."
      create_deployments --all
      
      echo "Waiting for pod to be ready..."
      wait_pod --image-selector ghcr.io/loft-sh/devspace-quickstart-golang --timeout 120
      
      echo "Running tests in container..."
      exec_container --image-selector ghcr.io/loft-sh/devspace-quickstart-golang -- go test ./...
      
      echo "Tests complete!"
```

### Step 2: Run the Pipeline

```bash
devspace run-pipeline test
```

DevSpace:
1. Deploys the app
2. Waits for the pod to be ready
3. Runs `go test ./...` inside the container
4. Prints "Tests complete!"

This is useful for integration tests that need the app running in Kubernetes.

## Hands-On: Add Custom Flags

Custom flags let users pass options to your pipelines.

### Step 1: Define Flags

Add a `flags` section to a pipeline:

```yaml
pipelines:
  deploy:
    flags:
    - name: skip-tests
      short: s
      type: bool
      default: false
      description: "Skip running tests before deploy"
    - name: tag
      short: t
      type: string
      description: "Custom image tag"
    run: |-
      SKIP_TESTS=$(get_flag "skip-tests")
      TAG=$(get_flag "tag")
      
      if ! is_true $SKIP_TESTS; then
        echo "Running tests..."
        go test ./...
      else
        echo "Skipping tests (--skip-tests flag set)"
      fi
      
      run_dependencies --all
      
      if ! is_empty $TAG; then
        build_images --all -t $TAG
      else
        build_images --all -t $(git describe --always)
      fi
      
      create_deployments --all
```

### Step 2: Use the Flags

**Skip tests:**

```bash
devspace deploy --skip-tests
# Or: devspace deploy -s
```

**Custom tag:**

```bash
devspace deploy --tag v1.2.3
# Or: devspace deploy -t v1.2.3
```

**Both:**

```bash
devspace deploy -s -t v1.2.3
```

DevSpace skips tests and tags images with `v1.2.3`.

## Advanced: Parallel Execution

You can run pipelines in parallel with `run_pipelines --background`.

### Example: Run Tests and Build in Parallel

```yaml
pipelines:
  ci:
    run: |-
      run_pipelines lint test --background
      echo "Lint and test complete. Building..."
      build_images --all
      create_deployments --all
  
  lint:
    run: |-
      echo "Running linter..."
      golangci-lint run ./...
  
  test:
    run: |-
      echo "Running tests..."
      go test ./...
```

Run:

```bash
devspace run-pipeline ci
```

`lint` and `test` pipelines run in parallel, then the build continues.

## Advanced: Conditional Logic

Use shell conditionals (`if`, `&&`, `||`) in pipelines.

### Example: Deploy Only on Main Branch

```yaml
pipelines:
  deploy:
    run: |-
      if ! is_equal ${DEVSPACE_GIT_BRANCH} "main"; then
        echo "Not on main branch. Skipping deploy."
        exit 0
      fi
      
      echo "Deploying from main branch..."
      run_dependencies --all
      build_images --all -t $(git describe --always)
      create_deployments --all
```

Run:

```bash
devspace deploy
```

If you're not on the `main` branch, deployment is skipped.

## What You Learned

- **Pipelines** are POSIX shell scripts that define what happens for each DevSpace command
- **Built-in functions**: `build_images`, `create_deployments`, `start_dev`, `run_pipelines`, `exec_container`, etc.
- **Global helpers**: `is_equal`, `is_empty`, `get_flag`, `is_true`
- You can **customize default pipelines** (dev, deploy, build, purge) or create **custom pipelines**
- Run custom pipelines with `devspace run-pipeline <name>`
- **Custom flags** let users pass options to pipelines
- Use **conditional logic** (`if`, `&&`, `||`) to control pipeline flow

## Next Steps

You now know how to build flexible workflows with pipelines. Let's scale up to multi-service apps. Move on to [Chapter 8: Dependencies](../08-dependencies/README.md).

## Troubleshooting

**Problem: Pipeline syntax error**
- Check for missing quotes, especially in `$(...)` expressions
- Make sure you use `if ... ; then` (note the semicolon)
- Validate YAML syntax (proper indentation)

**Problem: Function not found**
- Make sure you're using a pipeline-only function (`build_images`, `create_deployments`) inside a pipeline
- Global functions (`is_equal`, `is_empty`) can be used anywhere

**Problem: Custom flag not working**
- Make sure you defined the flag in the `flags` section
- Use `get_flag "flag-name"` to read it (quotes around the name)
- Check flag type (bool, string, int, stringArray)

**Problem: Pipeline runs but nothing happens**
- Add `echo` statements to debug: `echo "Building images..."`
- Check if conditionals are skipping steps
- Run `devspace print` to see if the pipeline is defined correctly
