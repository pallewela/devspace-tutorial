# Chapter 11: Troubleshooting and Best Practices

## Learning Objectives

By the end of this chapter, you will:
- Know how to debug common DevSpace issues
- Understand useful DevSpace commands for troubleshooting
- Learn best practices for organizing and maintaining `devspace.yaml`
- Know where to find help and documentation

## Prerequisites

- Completed [Chapter 10: CI/CD and Cleanup](../10-cicd-cleanup/README.md)

## Common Issues and Solutions

### Issue: Wrong Kubernetes Context or Namespace

**Symptoms:**
- DevSpace deploys to the wrong cluster
- `kubectl get pods` shows pods in the wrong namespace
- Permission errors

**Solution:**

Check your current context and namespace:

```bash
kubectl config current-context
devspace list namespaces
```

Switch to the correct context:

```bash
devspace use context
```

Select your context from the list.

Switch to the correct namespace:

```bash
devspace use namespace <your-namespace>
```

**Prevention:**
- Always run `devspace use context` and `devspace use namespace` when switching projects or clusters.
- Add a reminder to your README or onboarding docs.

### Issue: Image Pull Errors (ErrImagePull, ImagePullBackOff)

**Symptoms:**
- Pod stuck in `ErrImagePull` or `ImagePullBackOff` status
- `kubectl describe pod` shows "Failed to pull image"

**Root Causes:**

1. **Image doesn't exist** (typo in image name, or image wasn't built/pushed)
2. **Registry is private and there's no pull secret**
3. **Image was built locally but cluster can't access it**

**Solution:**

**For local clusters (minikube, kind):**
- Make sure you selected "Skip Registry" during `devspace init`.
- For minikube: Build images in minikube's Docker daemon: `eval $(minikube docker-env)` then `devspace build`.
- For kind: Load images into kind: `kind load docker-image <image> --name <cluster-name>`.

**For remote clusters:**
- Make sure the image was pushed: Check your registry UI or run `docker pull <image>`.
- Make sure the cluster can access the registry (firewall rules, private networks, etc.).
- If the registry is private, create a pull secret:

```bash
kubectl create secret docker-registry my-registry-secret \
  --docker-server=<registry-url> \
  --docker-username=<username> \
  --docker-password=<password> \
  -n <namespace>
```

Then add it to your deployment:

```yaml
deployments:
  app:
    helm:
      values:
        pullSecrets:
        - my-registry-secret
```

Or let DevSpace create it automatically (DevSpace prompts you during `devspace deploy`).

### Issue: File Sync Is Slow or Not Working

**Symptoms:**
- Files don't sync when you edit them locally
- Sync takes a long time
- "Sync: ↑ <file>" messages don't appear

**Solutions:**

1. **Exclude large directories:**

```yaml
sync:
- path: ./
  excludePaths:
  - .git/
  - node_modules/
  - vendor/
  - dist/
  - .devspace/
```

2. **Check if the file is excluded:**
   - Look at your `excludePaths` and make sure the file isn't matched.

3. **Restart dev mode:**
   - `Ctrl+C` to stop `devspace dev`, then run `devspace dev` again.

4. **Check container file system:**

```bash
devspace enter
ls -la /path/to/synced/files
```

5. **Check network latency** (for remote clusters):
   - File sync is slower over the internet. Consider using a VPN or a closer cluster.

### Issue: Build Fails

**Symptoms:**
- `devspace deploy` or `devspace build` fails with a build error
- Docker build fails

**Common Causes:**

1. **Dockerfile syntax error**
2. **Missing files** (`.dockerignore` excluded something needed)
3. **Docker daemon not running**
4. **Build context is too large** (slow or timeout)

**Solutions:**

1. **Test the Dockerfile manually:**

```bash
docker build -t test .
```

If it fails, fix the Dockerfile.

2. **Check Docker daemon:**

```bash
docker ps
```

If it fails, start Docker.

3. **Add a `.dockerignore` file** to exclude unnecessary files:

```
.git
.devspace
node_modules
dist
*.log
```

4. **Use a smaller build context:**

```yaml
images:
  app:
    context: ./src  # Only send ./src to Docker
```

### Issue: Deployment Fails

**Symptoms:**
- `devspace deploy` fails with a Helm or kubectl error
- Pods are in `CrashLoopBackOff` or `Error` status

**Solutions:**

1. **Check Helm values:**

```bash
devspace print
```

Look at the `deployments` section and verify the values are correct.

2. **Check pod logs:**

```bash
kubectl logs <pod-name> -n <namespace>
```

Or:

```bash
devspace logs
```

3. **Describe the pod:**

```bash
kubectl describe pod <pod-name> -n <namespace>
```

Look for errors in the "Events" section.

4. **Check resource limits:**
   - If the pod is `OOMKilled`, increase memory limits.
   - If the pod is `CrashLoopBackOff`, the app might be failing to start. Check logs.

### Issue: Pipeline Fails

**Symptoms:**
- `devspace deploy` or `devspace dev` fails with a pipeline error
- Custom pipeline doesn't work as expected

**Solutions:**

1. **Add debug output:**

```yaml
pipelines:
  deploy:
    run: |-
      echo "Starting deploy pipeline..."
      echo "Building images..."
      build_images --all
      echo "Deploying..."
      create_deployments --all
      echo "Deploy complete!"
```

2. **Check for syntax errors:**
   - Make sure you use `if ... ; then ... fi` (note semicolons).
   - Quote variables: `"${VAR}"`.

3. **Test functions individually:**

```bash
devspace build  # Test build_images
devspace deploy --skip-build  # Skip build, test deploy only
```

4. **Check exit codes:**
   - If a command fails (non-zero exit code), the pipeline stops.
   - Use `|| true` to ignore errors: `go test ./... || true`.

### Issue: Variables Not Working

**Symptoms:**
- Variable isn't replaced in the config
- DevSpace keeps asking for a variable value

**Solutions:**

1. **Use correct syntax:**
   - Use `${VAR_NAME}`, not `$VAR_NAME` or `$VAR_NAME` (braces are required).

2. **Check if the variable is defined:**

```bash
devspace list vars
```

3. **View the resolved config:**

```bash
devspace print
```

Look for `${VAR_NAME}` (not replaced) vs `actual-value` (replaced).

4. **Reset the variable cache:**

```bash
devspace reset vars
```

DevSpace will ask for the variable again.

5. **Pass the variable explicitly:**

```bash
devspace deploy --var VAR_NAME=value
```

## Useful Commands

### `devspace print`

Print the fully resolved config (after variables and profiles are applied):

```bash
devspace print
devspace print -p production  # With profile
```

**Use case:** Debug config issues, see what DevSpace sees.

### `devspace list vars`

List all variables and their current values:

```bash
devspace list vars
```

### `devspace reset vars`

Reset the variable cache (force DevSpace to ask for variables again):

```bash
devspace reset vars
```

### `devspace list profiles`

List all defined profiles:

```bash
devspace list profiles
```

### `devspace logs`

Stream logs from pods:

```bash
devspace logs
devspace logs --follow  # Keep streaming (like tail -f)
```

### `devspace enter`

Open a shell in a container:

```bash
devspace enter
```

**Use case:** Debug inside the container, run commands, inspect files.

### `devspace ui`

Open the DevSpace UI (even if dev mode isn't running):

```bash
devspace ui
```

### `devspace purge`

Remove deployed resources:

```bash
devspace purge
devspace purge --force  # Skip confirmation
```

### `devspace build`

Build images without deploying:

```bash
devspace build
devspace build --force-rebuild  # Force rebuild even if nothing changed
```

### `devspace deploy`

Deploy without starting dev mode:

```bash
devspace deploy
devspace deploy -p production  # With profile
devspace deploy --var VAR=value  # Override variable
```

### `devspace analyze`

Analyze your cluster and identify issues (pod failures, resource limits, etc.):

```bash
devspace analyze
```

**Use case:** Quick health check of your deployments.

## Best Practices

### 1. Version Your devspace.yaml with Code

Always commit `devspace.yaml` to git. This ensures:
- Everyone on the team uses the same config.
- You can see config changes in pull requests.
- You can roll back config changes if something breaks.

### 2. Use Variables and Profiles

Don't hardcode environment-specific values. Use variables and profiles:

- **Variables**: For values that differ per developer (ports, image registries).
- **Profiles**: For values that differ per environment (dev, staging, production).

**Example:**

```yaml
vars:
  IMAGE_REGISTRY:
    source: env
    default: docker.io/myusername

profiles:
  production:
    merge:
      images:
        app:
          image: gcr.io/my-prod-project/myapp
```

### 3. Use devImage + Sync in Dev Mode

Don't rebuild images during `devspace dev`. Use a `devImage` with dev tools and sync your code.

**Why:**
- Rebuilding is slow (1-2 minutes per rebuild).
- Syncing is instant (< 1 second).

**How:**

```yaml
images:
  app:
    image: my-registry/myapp
    dockerfile: ./Dockerfile

dev:
  app:
    devImage: my-registry/myapp-dev  # Includes dev tools
    sync:
    - path: ./
```

### 4. Exclude Unnecessary Files from Sync

Don't sync everything. Exclude:
- Version control (`.git/`)
- Dependencies (`node_modules/`, `vendor/`)
- Build artifacts (`dist/`, `build/`)
- Logs (`*.log`)
- DevSpace cache (`.devspace/`)

```yaml
sync:
- path: ./
  excludePaths:
  - .git/
  - node_modules/
  - dist/
  - '*.log'
  - .devspace/
```

### 5. Document Custom Pipelines and Commands

If you define custom pipelines or commands, document them in your README:

```markdown
## Custom Commands

- `devspace run migrate` - Run database migrations
- `devspace run test` - Run tests in the container
- `devspace run seed` - Seed the database with test data
```

### 6. Use Helm for Deployments (Recommended)

Helm is more flexible than raw kubectl manifests:
- Easier to parameterize (values can be overridden).
- Easier to manage releases (rollback, history, etc.).
- The component-chart covers most use cases.

If you need advanced Kubernetes features, you can still use `kubectl` or `kustomize`.

### 7. Test Profiles Locally Before CI

Before pushing a profile to CI, test it locally:

```bash
devspace deploy -p staging
devspace print -p production  # Verify the config looks right
```

### 8. Use `devspace purge` for Ephemeral Environments

For dev or feature-branch environments, clean up when you're done:

```bash
devspace purge
```

This frees up cluster resources.

For production, be more careful. Consider:
- Manual approval for purge (e.g., in GitLab CI: `when: manual`).
- Disabling purge entirely in production profiles.

### 9. Keep devspace.yaml Simple

Don't over-engineer your config. Start simple and add complexity only when needed:
- **Simple**: Single image, single deployment, basic sync.
- **Add as needed**: Multiple images, dependencies, custom pipelines, profiles.

### 10. Use the Official Docs

DevSpace has excellent documentation:
- **Getting Started**: [https://devspace.sh/docs/getting-started/introduction](https://devspace.sh/docs/getting-started/introduction)
- **Configuration Reference**: [https://devspace.sh/docs/configuration/reference](https://devspace.sh/docs/configuration/reference)
- **Pipelines**: [https://devspace.sh/docs/configuration/pipelines](https://devspace.sh/docs/configuration/pipelines)
- **File Sync**: [https://devspace.sh/docs/configuration/dev/connections/file-sync](https://devspace.sh/docs/configuration/dev/connections/file-sync)
- **Dependencies**: [https://devspace.sh/docs/configuration/dependencies](https://devspace.sh/docs/configuration/dependencies)

## Where to Get Help

### 1. Official Documentation

[https://devspace.sh/docs](https://devspace.sh/docs)

Comprehensive guides, config reference, and examples.

### 2. GitHub Issues

[https://github.com/loft-sh/devspace/issues](https://github.com/loft-sh/devspace/issues)

Report bugs, request features, or search for existing issues.

### 3. Slack Community

[https://slack.loft.sh/](https://slack.loft.sh/)

Chat with the DevSpace team and community. Fast responses to questions.

### 4. Stack Overflow

Tag your questions with `devspace` or `kubernetes`.

## What You Learned

- **Common issues**: Wrong context/namespace, image pull errors, slow sync, build failures, deployment failures, pipeline errors, variable issues.
- **Useful commands**: `devspace print`, `devspace list vars`, `devspace reset vars`, `devspace logs`, `devspace enter`, `devspace analyze`.
- **Best practices**: Version your config, use variables and profiles, use devImage + sync in dev, exclude unnecessary files, document custom pipelines, test profiles locally, use `devspace purge` for ephemeral environments.
- **Where to get help**: Official docs, GitHub issues, Slack, Stack Overflow.

## Congratulations!

You've completed the DevSpace tutorial. You now know:
- How to set up DevSpace and a Kubernetes cluster
- How to initialize and deploy projects
- How to use development mode with hot reloading
- How to configure images, variables, profiles, and pipelines
- How to compose multi-service apps with dependencies
- How to integrate with your IDE and use DevSpace in CI/CD
- How to troubleshoot common issues and follow best practices

**Next Steps:**
- Apply DevSpace to your own projects.
- Explore advanced features (custom build engines, hooks, advanced profiles).
- Join the DevSpace community on Slack.
- Contribute to DevSpace (it's open source!).

Thank you for following this tutorial. Happy coding with DevSpace!

## Troubleshooting This Tutorial

**Problem: Examples don't work**
- Make sure you're using DevSpace 6.x: `devspace version`
- Make sure you're in the correct directory (e.g., `devspace-quickstart-golang` for Chapters 2-7, `samples/multi-service` for Chapter 8).
- Check if your cluster is running: `kubectl get nodes`
- Check the DevSpace logs for errors: `devspace deploy --debug`

**Problem: Can't clone the official quickstart**
- Make sure git is installed: `git --version`
- Clone manually: `git clone https://github.com/loft-sh/devspace-quickstart-golang`
- If you get a 404, the repo might have moved. Check [https://github.com/loft-sh](https://github.com/loft-sh) for quickstarts.
