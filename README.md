# DevSpace Tutorial

A multi-chapter, hands-on tutorial that takes you from zero to deep [DevSpace](https://www.devspace.sh/) proficiency. No prior Kubernetes knowledge is required—we explain concepts when you need them.

DevSpace is an open-source, client-only developer tool for Kubernetes. It lets you **build, deploy, and develop** cloud-native apps with a single config file, hot reloading, and the same workflow locally and in CI.

## What You'll Learn

- Set up DevSpace and a local Kubernetes cluster
- Initialize a project, deploy it, and understand `devspace.yaml`
- Use development mode with file sync, port forwarding, and the DevSpace UI
- Configure image builds, variables, profiles, and pipelines
- Compose multi-service apps with dependencies
- Integrate with your IDE and run DevSpace in CI/CD

## Prerequisites

- **DevSpace CLI** (6.x) — [Installation](https://devspace.sh/cli/docs/getting-started/installation)
- **Docker** — for building images (and for minikube/kind if you use them)
- **kubectl** — [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- A **Kubernetes cluster** — local (minikube, kind, or k3s) or cloud (GKE, EKS, AKS, etc.)

Chapter 1 walks you through installing these and creating a local cluster if you don't have one.

## Table of Contents

### Part A: Get Started

| Chapter | Topic | Description |
|--------|--------|-------------|
| [1. Introduction and Environment Setup](chapters/01-setup/README.md) | Setup | What is DevSpace, install CLI and cluster, verify access |
| [2. Your First DevSpace Project](chapters/02-first-project/README.md) | Init & Deploy | Clone a quickstart, run `devspace init`, then `devspace deploy` |
| [3. Understanding devspace.yaml](chapters/03-understanding-config/README.md) | Config | Anatomy of pipelines, images, deployments, and dev |

### Part B: Development Workflow

| Chapter | Topic | Description |
|--------|--------|-------------|
| [4. Development Mode (devspace dev)](chapters/04-development-mode/README.md) | Dev workflow | File sync, port forwarding, DevSpace UI, hot reload |
| [5. Images and Builds](chapters/05-images-builds/README.md) | Image config | Dockerfile, buildArgs, tagging, when images are built |

### Part C: Configuration and Pipelines

| Chapter | Topic | Description |
|--------|--------|-------------|
| [6. Variables and Profiles](chapters/06-variables-profiles/README.md) | Config flexibility | Vars (env, command, question), profiles for dev/staging/prod |
| [7. Pipelines](chapters/07-pipelines/README.md) | Custom workflows | Pipeline scripts, built-in functions, custom pipelines and flags |

### Part D: Multi-Service and Production

| Chapter | Topic | Description |
|--------|--------|-------------|
| [8. Dependencies](chapters/08-dependencies/README.md) | Multi-service | Path and git dependencies, referencing dependency images |
| [9. Dev Containers and IDE Integration](chapters/09-dev-containers-ide/README.md) | Dev containers | Sync, terminal, SSH, custom commands |
| [10. CI/CD and Cleanup](chapters/10-cicd-cleanup/README.md) | Production | Non-interactive CI, purge, example GitHub Actions |

### Part E: Wrap-Up

| Chapter | Topic | Description |
|--------|--------|-------------|
| [11. Troubleshooting and Best Practices](chapters/11-troubleshooting/README.md) | Operations | Common issues, useful commands, best practices, next steps |

## Sample Code

- **[samples/minimal-app](samples/minimal-app/)** — Minimal Go app with a Dockerfile for Chapters 3, 5, and 6.
- **[samples/multi-service](samples/multi-service/)** — Two-service (API + frontend) example for Chapter 8.

Chapters 2 and 4 use the official [devspace-quickstart-golang](https://github.com/loft-sh/devspace-quickstart-golang) repo; you'll clone it when you run the tutorial.

## Consistency Note

This tutorial targets **DevSpace 6.x** and uses **minikube** as the default local cluster. If you prefer **kind** or **k3s**, Chapter 1 includes an alternative; the rest of the tutorial works the same.

## License and Links

- [DevSpace Documentation](https://devspace.sh/docs/getting-started/introduction)
- [DevSpace on GitHub](https://github.com/loft-sh/devspace)
- [DevSpace Slack](https://slack.loft.sh/)
