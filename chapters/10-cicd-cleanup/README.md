# Chapter 10: CI/CD and Cleanup

## Learning Objectives

By the end of this chapter, you will:
- Understand how to use DevSpace in CI/CD pipelines
- Configure DevSpace for non-interactive environments
- Use profiles and variables for staging/production deployments
- Write a GitHub Actions workflow that uses DevSpace
- Use `devspace purge` to clean up resources
- Know when and how to tear down dev and staging environments

## Prerequisites

- Completed [Chapter 9: Dev Containers and IDE Integration](../09-dev-containers-ide/README.md)
- Have the `devspace-quickstart-golang` project initialized

## Why Use DevSpace in CI/CD?

DevSpace isn't just for local development. The same `devspace.yaml` that works on your laptop works in CI:

- **One config**: Dev, staging, and production use the same base config (customized with profiles)
- **Reproducible**: What you deploy locally is what CI deploys
- **No duplicated logic**: You don't need separate CI scripts for building, tagging, and deploying

### Typical CI Workflow

1. **Checkout code** (e.g., `git clone` or `actions/checkout`)
2. **Install DevSpace CLI**
3. **Authenticate** to your Kubernetes cluster and image registry
4. **Run `devspace deploy`** with a profile (e.g., `-p staging` or `-p production`)
5. **Run tests** (optional, e.g., `devspace run-pipeline test`)
6. **Notify** (Slack, email, etc.)

## Non-Interactive Mode

CI environments don't have a TTY (terminal), so DevSpace runs in non-interactive mode automatically. This means:

- **No prompts**: Variables with questions won't prompt the user; they'll use the default or fail if no default is set.
- **No terminal**: The `terminal` in dev mode is disabled by default.
- **No colors**: Log output is plain text (no ANSI colors).

### Ensure Non-Interactive Variables

For CI, make sure all variables have defaults or are passed via `--var` or environment variables:

**Bad (prompts in CI):**

```yaml
vars:
  IMAGE_REGISTRY:
    question: Which registry?
```

**Good (has default or uses env var):**

```yaml
vars:
  IMAGE_REGISTRY:
    source: env
    default: docker.io/myusername
```

Or pass it explicitly:

```bash
devspace deploy --var IMAGE_REGISTRY=gcr.io/my-project
```

## Using Profiles for Staging/Production

Define profiles for different environments.

### Example: Staging Profile

```yaml
profiles:
  staging:
    merge:
      images:
        app:
          image: gcr.io/my-project-staging/myapp
      deployments:
        app:
          helm:
            values:
              replicas: 2
              resources:
                limits:
                  cpu: "1"
                  memory: "512Mi"
```

### Example: Production Profile

```yaml
profiles:
  production:
    merge:
      images:
        app:
          image: gcr.io/my-project-prod/myapp
      deployments:
        app:
          helm:
            values:
              replicas: 3
              resources:
                limits:
                  cpu: "2"
                  memory: "1Gi"
    replace:
      dev: {}  # Disable dev features in production
```

### Deploy to Staging

```bash
devspace deploy -p staging --var ENVIRONMENT=staging
```

### Deploy to Production

```bash
devspace deploy -p production --var ENVIRONMENT=production
```

## GitHub Actions Example

Let's create a GitHub Actions workflow that deploys to staging on push to `develop` and to production on push to `main`.

### Create `.github/workflows/deploy.yml`

```yaml
name: Deploy to Kubernetes

on:
  push:
    branches:
    - develop
    - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Install DevSpace
      run: |
        curl -sSL https://devspace.sh/install.sh | bash
    
    - name: Authenticate to Google Cloud
      uses: google-github-actions/auth@v1
      with:
        credentials_json: ${{ secrets.GCP_SA_KEY }}
    
    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
    
    - name: Configure kubectl
      run: |
        gcloud container clusters get-credentials my-cluster --region us-central1 --project my-project
    
    - name: Authenticate to Container Registry
      run: |
        gcloud auth configure-docker gcr.io
    
    - name: Determine environment
      id: env
      run: |
        if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
          echo "profile=production" >> $GITHUB_OUTPUT
          echo "namespace=production" >> $GITHUB_OUTPUT
        else
          echo "profile=staging" >> $GITHUB_OUTPUT
          echo "namespace=staging" >> $GITHUB_OUTPUT
        fi
    
    - name: Deploy with DevSpace
      run: |
        devspace use namespace ${{ steps.env.outputs.namespace }}
        devspace deploy -p ${{ steps.env.outputs.profile }} --var IMAGE_TAG=${{ github.sha }}
      env:
        DEVSPACE_FLAGS: "-s"  # Silent mode (less verbose)
    
    - name: Notify Slack
      if: success()
      run: |
        curl -X POST -H 'Content-type: application/json' \
          --data '{"text":"Deployed to ${{ steps.env.outputs.namespace }} successfully!"}' \
          ${{ secrets.SLACK_WEBHOOK_URL }}
```

### What This Does

1. **Checkout**: Clone the repo.
2. **Install DevSpace**: Download and install the CLI.
3. **Authenticate to GCP**: Use a service account key (stored in GitHub Secrets).
4. **Configure kubectl**: Point kubectl to your GKE cluster.
5. **Authenticate to GCR**: Allow Docker to push images.
6. **Determine environment**: If branch is `main`, use production profile; otherwise, staging.
7. **Deploy**: Run `devspace deploy` with the appropriate profile.
8. **Notify Slack**: Send a success message (optional).

### Secrets to Set in GitHub

- `GCP_SA_KEY`: Google Cloud service account JSON key (with permissions to deploy to GKE and push to GCR)
- `SLACK_WEBHOOK_URL`: Slack webhook for notifications (optional)

## GitLab CI Example

### Create `.gitlab-ci.yml`

```yaml
image: ubuntu:22.04

stages:
- deploy

variables:
  DEVSPACE_FLAGS: "-s"

before_script:
- apt-get update && apt-get install -y curl
- curl -sSL https://devspace.sh/install.sh | bash
- export PATH=$PATH:$HOME/.devspace/bin

deploy-staging:
  stage: deploy
  only:
  - develop
  script:
  - echo "$KUBECONFIG_STAGING" > kubeconfig.yaml
  - export KUBECONFIG=kubeconfig.yaml
  - devspace use namespace staging
  - devspace deploy -p staging --var IMAGE_TAG=$CI_COMMIT_SHA
  environment:
    name: staging
    url: https://staging.myapp.com

deploy-production:
  stage: deploy
  only:
  - main
  script:
  - echo "$KUBECONFIG_PRODUCTION" > kubeconfig.yaml
  - export KUBECONFIG=kubeconfig.yaml
  - devspace use namespace production
  - devspace deploy -p production --var IMAGE_TAG=$CI_COMMIT_SHA
  environment:
    name: production
    url: https://myapp.com
  when: manual  # Require manual approval for production
```

### GitLab CI Variables to Set

- `KUBECONFIG_STAGING`: Base64-encoded kubeconfig for staging cluster
- `KUBECONFIG_PRODUCTION`: Base64-encoded kubeconfig for production cluster

## Using `devspace purge`

`devspace purge` removes deployed resources from the cluster. It runs the `purge` pipeline.

### Default Purge Pipeline

```yaml
pipelines:
  purge:
    run: |-
      stop_dev --all
      purge_deployments --all
      run_dependencies --all --pipeline purge
```

**What happens:**

1. **`stop_dev --all`**: Stop all running dev sessions.
2. **`purge_deployments --all`**: Delete all Helm releases or kubectl resources.
3. **`run_dependencies --all --pipeline purge`**: Purge dependencies (if any).

### Run Purge

```bash
devspace purge
```

DevSpace asks for confirmation (unless you use `--force`).

**Verify:**

```bash
kubectl get all -n devspace-tutorial
```

You should see:

```
No resources found in devspace-tutorial namespace.
```

### When to Use Purge

- **Tear down dev environments**: When you're done working on a feature or branch.
- **Clean up staging**: After testing a release.
- **Reset your cluster**: If something went wrong and you want a fresh start.

**Warning:** In production, use `purge` with care. You might want to disable it or require manual approval.

### Selective Purge

Purge only specific deployments:

```bash
devspace purge app  # Purge only the 'app' deployment
```

Or customize the purge pipeline:

```yaml
pipelines:
  purge:
    run: |-
      purge_deployments app frontend  # Don't purge database
```

## Hands-On: Deploy to a "Staging" Namespace

Let's simulate a staging deployment.

### Step 1: Create a Staging Profile

Open `devspace.yaml` and add:

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

### Step 2: Deploy to Staging Namespace

```bash
devspace use namespace staging
devspace deploy -p staging
```

DevSpace:
1. Creates the `staging` namespace (if it doesn't exist)
2. Deploys with 2 replicas

### Step 3: Verify

```bash
kubectl get pods -n staging
```

You should see 2 app pods.

### Step 4: Purge Staging

```bash
devspace purge
```

DevSpace removes the deployment from the `staging` namespace.

## Advanced: Blue-Green Deployments

You can use DevSpace for blue-green deployments with Helm:

1. Deploy the new version to a "green" namespace.
2. Run smoke tests.
3. Switch traffic from "blue" to "green" (e.g., update an ingress).
4. Purge the "blue" namespace.

This is beyond the scope of this tutorial, but DevSpace's pipelines and profiles make it easy to script.

## What You Learned

- DevSpace works in CI/CD pipelines with the same `devspace.yaml` used locally
- Use **profiles** to customize config for staging/production
- Pass **variables** via `--var` or environment variables (no prompts in CI)
- **GitHub Actions** and **GitLab CI** examples show how to integrate DevSpace
- `devspace purge` removes deployed resources (use with care in production)
- You can deploy to multiple namespaces and use profiles to control replicas, resources, and images

## Next Steps

You've learned how to use DevSpace from local dev to production. Let's wrap up with troubleshooting tips and best practices. Move on to [Chapter 11: Troubleshooting and Best Practices](../11-troubleshooting/README.md).

## Troubleshooting

**Problem: CI fails with "variable not defined"**
- Make sure all variables have defaults or are passed via `--var` or environment variables
- Check `devspace print` locally to see what variables are used

**Problem: CI can't push images**
- Make sure Docker is authenticated to your registry (e.g., `docker login`, `gcloud auth configure-docker`, `aws ecr get-login-password`)
- Check that the service account or CI runner has permissions to push

**Problem: CI can't connect to Kubernetes**
- Make sure `kubectl` is configured (e.g., `gcloud container clusters get-credentials`, `aws eks update-kubeconfig`)
- Check that the service account or CI runner has permissions to deploy to the cluster
- Test with `kubectl get nodes` in CI

**Problem: Purge fails with "dependency still in use"**
- Another project might depend on your deployment
- Use `--force-purge` to override (with care): `devspace purge --force-purge`
