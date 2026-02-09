# Chapter 1: Introduction and Environment Setup

## Learning Objectives

By the end of this chapter, you will:
- Understand what DevSpace is and why developers use it
- Have DevSpace CLI, Docker, and kubectl installed
- Have a working Kubernetes cluster (local or cloud)
- Be able to verify DevSpace can connect to your cluster

## What is DevSpace?

DevSpace is an open-source, **client-only** developer tool for Kubernetes. It runs as a single CLI binary on your local machine and communicates directly with your Kubernetes cluster using your kube-context (just like kubectl).

### Key Benefits

1. **Fast development workflow**: Hot reload your code without rebuilding images or restarting containers
2. **Single config file**: `devspace.yaml` defines build, deploy, and dev workflows—share it with your team via git
3. **Works anywhere**: Local clusters (minikube, kind, k3s) or cloud clusters (GKE, EKS, AKS, DOKS)
4. **No cluster-side install**: DevSpace doesn't require anything installed in your cluster
5. **Team-friendly**: DevOps defines the config; developers just run `devspace dev` or `devspace deploy`

### How It Works

```
┌─────────────────┐
│  Your Computer  │
│                 │
│  DevSpace CLI   │──────> Uses kubectl config
│  (binary)       │        and kube-context
└────────┬────────┘
         │
         │ Direct connection
         │ (no server-side component)
         v
┌─────────────────┐
│ Kubernetes API  │
│   (cluster)     │
└─────────────────┘
```

DevSpace reads your local `devspace.yaml`, then builds images, deploys manifests/Helm charts, and starts development features (file sync, port forwarding, terminal) directly.

## Prerequisites

You'll need three things:
1. **DevSpace CLI** (version 6.x)
2. **Docker** (for building images and optionally running local clusters)
3. **kubectl** (for interacting with Kubernetes)
4. **A Kubernetes cluster** (we'll help you set up a local one if you don't have one)

### 1. Install DevSpace CLI

**macOS:**
```bash
curl -sSL https://devspace.sh/install.sh | bash
```

**Linux:**
```bash
curl -sSL https://devspace.sh/install.sh | bash
```

**Windows (PowerShell):**
```powershell
md -Force "$Env:APPDATA\devspace"; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]'Tls,Tls11,Tls12';
Invoke-WebRequest -UseBasicParsing ((Invoke-WebRequest -URI "https://github.com/loft-sh/devspace/releases/latest" -UseBasicParsing).Content -replace "(?ms).*`"([^`"]*devspace-windows-amd64.exe)`".*","https://github.com/`$1") -o $Env:APPDATA\devspace\devspace.exe;
$env:Path += ";" + $Env:APPDATA + "\devspace";
[Environment]::SetEnvironmentVariable("Path", $env:Path, [System.EnvironmentVariableTarget]::User);
```

**Verify installation:**
```bash
devspace version
```

You should see output like:
```
devspace version 6.x.x
```

### 2. Install Docker

Docker is needed for:
- Building container images
- Running local clusters (minikube with Docker driver, or kind)

**Download and install:** [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)

**Verify:**
```bash
docker --version
```

### 3. Install kubectl

kubectl is the Kubernetes command-line tool. DevSpace uses your kubectl config under the hood.

**Installation:** [https://kubernetes.io/docs/tasks/tools/](https://kubernetes.io/docs/tasks/tools/)

**Quick install commands:**

**macOS (Homebrew):**
```bash
brew install kubectl
```

**Linux:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

**Windows (PowerShell):**
```powershell
# Use Chocolatey
choco install kubernetes-cli
# Or download from https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/
```

**Verify:**
```bash
kubectl version --client
```

## Setting Up a Kubernetes Cluster

DevSpace works with **any** Kubernetes cluster. If you already have access to a cloud cluster (GKE, EKS, AKS, etc.), you can skip to the [Verify Cluster Access](#verify-cluster-access) section.

For local development, we recommend **minikube** or **kind**. Both are free, run on your machine, and work great with DevSpace.

### Option A: minikube (Recommended)

Minikube runs a single-node Kubernetes cluster in a VM or container.

**Install minikube:** [https://minikube.sigs.k8s.io/docs/start/](https://minikube.sigs.k8s.io/docs/start/)

**Start a cluster:**
```bash
minikube start --driver=docker
```

This will:
1. Download the Kubernetes components
2. Start a cluster using Docker
3. Configure kubectl to use this cluster automatically

**Verify:**
```bash
kubectl get nodes
```

You should see:
```
NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane   1m    v1.28.3
```

### Option B: kind (Kubernetes IN Docker)

kind runs Kubernetes clusters in Docker containers. It's fast and lightweight.

**Install kind:** [https://kind.sigs.k8s.io/docs/user/quick-start/](https://kind.sigs.k8s.io/docs/user/quick-start/)

**Create a cluster:**
```bash
kind create cluster --name devspace-tutorial
```

**Verify:**
```bash
kubectl get nodes
```

You should see:
```
NAME                             STATUS   ROLES           AGE   VERSION
devspace-tutorial-control-plane   Ready    control-plane   1m    v1.28.0
```

### Option C: Cloud Cluster (GKE, EKS, AKS, etc.)

If you're using a cloud cluster, make sure:
1. You have kubectl configured to access it (usually via `gcloud`, `aws`, or `az` CLI)
2. You can run `kubectl get nodes` successfully
3. You have permissions to create namespaces and deploy workloads

## Verify Cluster Access

Let's make sure DevSpace can see your cluster.

### 1. Check kubectl context

```bash
kubectl config current-context
```

This shows which cluster kubectl (and DevSpace) will talk to. For minikube, it's usually `minikube`. For kind, it's `kind-devspace-tutorial` (or whatever you named it).

### 2. List available contexts

```bash
kubectl config get-contexts
```

You'll see all configured clusters. The current one has a `*` next to it.

### 3. Tell DevSpace which context to use

```bash
devspace use context
```

DevSpace will show you a list of contexts. Select the one you want (e.g., `minikube` or `kind-devspace-tutorial`).

### 4. Create and select a namespace

Namespaces isolate resources in Kubernetes. Let's create one for this tutorial:

```bash
devspace use namespace devspace-tutorial
```

If the namespace doesn't exist, DevSpace will ask if you want to create it—say **yes**.

**Verify:**
```bash
kubectl get namespace devspace-tutorial
```

You should see:
```
NAME                STATUS   AGE
devspace-tutorial   Active   10s
```

### 5. Verify DevSpace can see the cluster

Run:
```bash
devspace list namespaces
```

You should see a list of namespaces in your cluster, including `devspace-tutorial`.

## Concepts: Just-in-Time Kubernetes Basics

You don't need to be a Kubernetes expert to use DevSpace, but here are a few terms you'll see in this tutorial:

- **Cluster**: A set of machines running Kubernetes. Can be one local machine (minikube/kind) or many cloud VMs.
- **Namespace**: A virtual cluster inside the cluster. Think of it like a folder for your apps. Isolates resources.
- **kube-context**: A configuration that tells kubectl (and DevSpace) which cluster and user to use.
- **Pod**: The smallest deployable unit in Kubernetes. Usually runs one container. Think of it as a running instance of your app.
- **Image**: A Docker/OCI container image. Your app code packaged as a container.

DevSpace hides most of the Kubernetes complexity, but these terms will come up.

## What You Learned

- DevSpace is a client-only CLI tool that streamlines Kubernetes development
- You installed DevSpace, Docker, and kubectl
- You set up a local Kubernetes cluster (or confirmed access to a cloud cluster)
- You selected a kube-context and namespace for DevSpace to use
- DevSpace can now see your cluster and is ready to deploy apps

## Next Steps

Now that your environment is ready, move on to [Chapter 2: Your First DevSpace Project](../02-first-project/README.md) where you'll initialize a real app, deploy it, and see DevSpace in action.

## Troubleshooting

**Problem: `devspace version` not found**
- Make sure the DevSpace binary is in your PATH. Re-run the install script or add the binary location to your PATH manually.

**Problem: `minikube start` fails**
- Make sure Docker is running: `docker ps`
- Try a different driver: `minikube start --driver=virtualbox` (requires VirtualBox)

**Problem: kubectl can't connect**
- For minikube: Run `minikube status` to check if the cluster is running
- For kind: Run `kind get clusters` to list clusters
- For cloud: Re-authenticate with your cloud CLI (gcloud, aws, az)

**Problem: Permission denied errors**
- Make sure your kube-context has permissions to create namespaces and deploy workloads
- For minikube/kind, you're the admin by default
