# Chapter 6: Variables and Profiles

## Learning Objectives

By the end of this chapter, you will:
- Understand why config variables are useful
- Use built-in variables (like `DEVSPACE_NAMESPACE`, `DEVSPACE_GIT_COMMIT`)
- Define custom variables (static, from env, from commands, from questions)
- Use variables in images, deployments, and dev config
- Create profiles to support different environments (dev, staging, production)
- Apply profiles with `-p` and see the resulting config with `devspace print`

## Prerequisites

- Completed [Chapter 5: Images and Builds](../05-images-builds/README.md)
- Have the `devspace-quickstart-golang` project initialized

## Why Variables?

Config variables make your `devspace.yaml` **dynamic and reusable**. Instead of hardcoding values, you use variables that can change per developer, per environment, or per CI run.

**Use cases:**
- Different ports per developer (Alice uses 8080, Bob uses 3000)
- Different image registries (dev vs production)
- Different namespaces (dev, staging, prod)
- Different replicas or resource limits (dev: 1 replica, prod: 3 replicas)
- CI-specific settings (disable terminal, use specific tags)

## Built-In Variables

DevSpace provides several built-in variables. You can use them anywhere in `devspace.yaml`.

### Common Built-In Variables

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `${DEVSPACE_NAMESPACE}` | Current namespace | `devspace-tutorial` |
| `${DEVSPACE_CONTEXT}` | Current kube-context | `minikube` |
| `${DEVSPACE_NAME}` | Project name (from `name` field) | `devspace-quickstart-golang` |
| `${DEVSPACE_GIT_COMMIT}` | Latest git commit hash | `abc1234` |
| `${DEVSPACE_GIT_BRANCH}` | Current git branch | `main` |
| `${DEVSPACE_RANDOM}` | Random 6-character string | `xyz789` |
| `${DEVSPACE_TIMESTAMP}` | UNIX timestamp | `1634567890` |
| `${DEVSPACE_TMPDIR}` | Temp directory (cleaned after run) | `/tmp/devspace-xyz` |

### Using Built-In Variables

Example: Tag images with the git commit hash:

```yaml
pipelines:
  deploy:
    run: |-
      build_images --all -t ${DEVSPACE_GIT_COMMIT}
      create_deployments --all
```

Now when you run `devspace deploy`, images are tagged with the commit hash (e.g., `abc1234`).

## Custom Variables

You can define your own variables in the `vars` section.

### Static Value

Define a variable with a fixed value:

```yaml
vars:
  APP_PORT: "8080"
  ENVIRONMENT: "dev"
```

Use it anywhere:

```yaml
dev:
  app:
    ports:
    - port: "${APP_PORT}:${APP_PORT}"
```

### From Environment Variable

Load a variable from the user's environment:

```yaml
vars:
  IMAGE_REGISTRY:
    source: env
    default: "docker.io/myusername"
```

If the user has `IMAGE_REGISTRY=ghcr.io/alice` in their shell, DevSpace uses that. Otherwise, it uses the default.

**Use it:**

```yaml
images:
  app:
    image: ${IMAGE_REGISTRY}/myapp
```

### From Command

Run a command and capture its output:

```yaml
vars:
  GIT_TAG: $(git describe --tags --always)
```

Or the long form:

```yaml
vars:
  GIT_TAG:
    command: git
    args: ["describe", "--tags", "--always"]
```

**Use it:**

```yaml
pipelines:
  deploy:
    run: |-
      build_images --all -t ${GIT_TAG}
      create_deployments --all
```

### From User Input (Question)

Prompt the user for a value:

```yaml
vars:
  APP_PORT:
    question: Which port should the app use?
    default: "8080"
```

When you run `devspace deploy`, DevSpace asks:

```
? Which port should the app use? (8080)
```

The user can press Enter (use default) or type a different value.

#### With Options (Picker)

```yaml
vars:
  ENVIRONMENT:
    question: Which environment are you deploying to?
    options:
    - dev
    - staging
    - production
    default: dev
```

DevSpace shows a picker:

```
? Which environment are you deploying to?
  > dev
    staging
    production
```

## Hands-On: Add a Variable

Let's add a variable for the app port and use it.

### Step 1: Define the Variable

Open `devspace.yaml` and add a `vars` section (at the top level, before or after `pipelines`):

```yaml
vars:
  APP_PORT:
    question: Which port should the app listen on?
    default: "8080"
```

### Step 2: Use the Variable in Deployment

Update the deployment to use the variable:

```yaml
deployments:
  app:
    helm:
      values:
        containers:
        - image: ghcr.io/loft-sh/devspace-quickstart-golang
        service:
          ports:
          - port: ${APP_PORT}
```

### Step 3: Use the Variable in Dev

Update the dev config:

```yaml
dev:
  app:
    ports:
    - port: "2345"
    - port: "${APP_PORT}:8080"  # Forward localhost:APP_PORT to container:8080
    open:
    - url: http://localhost:${APP_PORT}
```

### Step 4: Deploy with the Variable

```bash
devspace deploy
```

DevSpace asks:

```
? Which port should the app listen on? (8080)
```

Press Enter (use 8080) or type `3000`.

If you type `3000`, DevSpace:
- Deploys the service on port 3000
- Forwards `localhost:3000` to `container:8080` in dev mode
- Opens `http://localhost:3000` in your browser

### Step 5: Override on the Command Line

You can skip the prompt by passing `--var`:

```bash
devspace deploy --var APP_PORT=9090
```

DevSpace uses `9090` without asking.

### Step 6: View Variables

List all variables:

```bash
devspace list vars
```

You'll see `APP_PORT` and its value.

Reset the variable cache (force DevSpace to ask again):

```bash
devspace reset vars
```

## Profiles

Profiles let you **override parts of the config** for different use cases. For example:
- **dev profile**: 1 replica, use local registry
- **staging profile**: 2 replicas, use staging registry
- **production profile**: 3 replicas, use production registry, disable dev features

Profiles are defined in the `profiles` section.

### Profile Strategies

DevSpace supports three strategies for modifying config:

1. **Replace**: Replace entire sections (e.g., replace all `deployments`)
2. **Merge**: JSON Merge Patch (RFC 7386) — useful for changing specific fields
3. **Patches**: JSON Patch (RFC 6902) — precise changes (add, remove, replace paths)

We'll focus on **replace** and **merge** (most common).

### Basic Profile: Replace

```yaml
profiles:
  production:
    replace:
      images:
        app:
          image: gcr.io/my-prod-project/myapp
```

When you run `devspace deploy -p production`, the `images.app.image` is replaced with the production registry.

### Basic Profile: Merge

```yaml
profiles:
  staging:
    merge:
      deployments:
        app:
          helm:
            values:
              replicas: 2
```

When you run `devspace deploy -p staging`, DevSpace merges `replicas: 2` into the existing deployment values.

## Hands-On: Add a Production Profile

Let's create a profile that:
- Disables dev mode (no sync, no terminal)
- Tags images with a semantic version tag
- Uses 3 replicas

### Step 1: Define the Profile

Add a `profiles` section to `devspace.yaml`:

```yaml
profiles:
  production:
    merge:
      deployments:
        app:
          helm:
            values:
              replicas: 3
    replace:
      dev: {}  # Disable dev config
```

### Step 2: Deploy with the Profile

```bash
devspace deploy -p production
```

DevSpace:
- Builds the image
- Deploys with 3 replicas
- Does not include dev features (because `dev` is replaced with `{}`)

### Step 3: View the Merged Config

Want to see what the config looks like with the profile applied?

```bash
devspace print -p production
```

This shows the final config after merging the profile. Look for `replicas: 3` in the output.

### Step 4: Add a Variable for Replicas

Let's make replicas configurable:

```yaml
vars:
  REPLICAS:
    question: How many replicas?
    default: "1"

profiles:
  production:
    merge:
      deployments:
        app:
          helm:
            values:
              replicas: ${REPLICAS}
```

Now:

```bash
devspace deploy -p production --var REPLICAS=5
```

Deploys with 5 replicas.

## Advanced: Multiple Profiles

You can apply multiple profiles:

```bash
devspace deploy -p staging -p us-west
```

Profiles are applied in order. Later profiles override earlier ones.

## Advanced: Conditional Profiles

Use `activation` to auto-apply profiles based on conditions:

```yaml
profiles:
  minikube:
    activation:
    - context: minikube
    replace:
      images:
        app:
          skipPush: true
```

When your kube-context is `minikube`, this profile automatically applies (skips pushing images).

## .env File for Variables

You can define variables in a `.env` file (similar to Docker Compose).

Create `.env`:

```
APP_PORT=3000
IMAGE_REGISTRY=ghcr.io/alice
```

Tell DevSpace to use it:

```yaml
vars:
  DEVSPACE_ENV_FILE: ".env"
```

Or set the environment variable:

```bash
export DEVSPACE_ENV_FILE=.env
devspace deploy
```

Variables in `.env` are loaded automatically.

## What You Learned

- **Built-in variables** (like `DEVSPACE_NAMESPACE`, `DEVSPACE_GIT_COMMIT`) are available everywhere
- **Custom variables** can be static, from env, from commands, or from questions
- Use variables in any config field: `${VAR_NAME}`
- Override variables on the command line: `--var VAR_NAME=value`
- **Profiles** let you customize config for different environments (dev, staging, prod)
- Apply profiles with `-p`: `devspace deploy -p production`
- View the final config with `devspace print -p <profile>`

## Next Steps

Now that your config is flexible, let's dive into pipelines to customize workflows. Move on to [Chapter 7: Pipelines](../07-pipelines/README.md).

## Troubleshooting

**Problem: Variable not replaced in config**
- Make sure you use `${VAR_NAME}`, not `$VAR_NAME` (needs braces)
- Run `devspace print` to see the resolved config
- Check if the variable is defined in `vars` or as an environment variable

**Problem: Profile doesn't apply**
- Make sure you use `-p`: `devspace deploy -p production`
- Check profile syntax (indentation, YAML validity)
- Run `devspace print -p production` to see what the profile does

**Problem: DevSpace keeps asking for a variable**
- Variables are cached. To reset: `devspace reset vars`
- Use `--var` to skip the prompt: `devspace deploy --var MY_VAR=value`
