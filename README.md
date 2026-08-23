# Tyk SRE Assignment

A Go-based Kubernetes SRE tool implementing five capabilities:

1. Deployment health monitoring
2. Kubernetes workload network isolation
3. Kubernetes API connectivity checking
4. Container image build and CI/CD
5. Helm-based Kubernetes deployment

The application can run locally using a kubeconfig or inside Kubernetes using in-cluster authentication.

---

## Table of Contents

- [Architecture](#architecture)
- [Repository Structure](#repository-structure)
- [Prerequisites](#prerequisites)
- [Run Locally](#run-locally)
- [Build the Container](#build-the-container)
- [Deploy to Kubernetes](#deploy-to-kubernetes)
- [API](#api)
- [Stories and Design Decisions](#stories-and-design-decisions)
- [Limitations](#limitations)

---

# Repository Structure

```text
.
├── golang/
│   ├── internal/
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── ...
│
├── helm/
│   └── tyk-sre-assignment/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│
├── .github/
│   └── workflows/
│
└── README.md
```

The Go application is kept under `golang/`.

The Helm chart is kept at repository level because it represents Kubernetes deployment configuration rather than application source code.

CI/CD configuration is kept under `.github/workflows`.

---

# Prerequisites

## Local development

- Go
- Docker
- kubectl
- Helm

## Kubernetes testing

- Kubernetes cluster
- kind for local Kubernetes testing

A Kubernetes CNI that supports and enforces NetworkPolicy is required to fully test the network isolation story.

---

# Running Locally

The application supports kubeconfig-based authentication when running outside Kubernetes.

Clone the repo and go in the project directory -

```bash
git clone https://github.com/kavishbaghel/tyk-sre-assignment.git
cd tyk-sre-assignment
```

## Build

```bash
cd golang
go mod download
go mod verify
go build -o tyk-sre-assignment .
```

## Run

```bash
./tyk-sre-assignment --kubeconfig ~/.kube/config --address ":8080"
```

## Verify

Application health:

```bash
curl http://localhost:8080/healthz
```

Kubernetes API connectivity:

```bash
curl http://localhost:8080/connectivity
```

Deployment health:

```bash
curl http://localhost:8080/deployments/health
```


## Go Tests

Run the complete test suite:

```bash
cd golang
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Check formatting:

```bash
gofmt -l .
```

Expected:

```text
go test ./...    -> PASS
go vet ./...     -> no issues
gofmt -l .       -> no output
```

---

# Deploy to Kubernetes

The application is packaged as a Helm chart:

```text
helm/tyk-sre-assignment-app
```

The chart deploys the application into a dedicated namespace.

## 1. Create a kind cluster

```bash
kind create cluster --name tyk-sre
```

Verify the cluster details:

```bash
kubectl cluster-info --context kind-tyk-sre
```

## 2. Build the docker image and load the image into kind

If you are building from the repository root use the following command -

```bash
docker build -f golang/Dockerfile -t tyk-sre-assignment:local ./golang
```

If you are inside the `golang/` directory use the following command -

```bash
docker build -t tyk-sre-assignment:local .
```
Verify the image has been created -

```bash
docker image ls | grep "tyk-sre-assignment:local"
```

Load the image created locally in the Kind cluster.

```bash
kind load docker-image tyk-sre-assignment:local --name tyk-sre
```


## 4. Install the Helm chart

```bash
helm upgrade --install tyk-sre-assignment ./helm/tyk-sre-assignment-app \
  --set image.repository=tyk-sre-assignment \
  --set image.tag=local
```

## 5. Verify the Deployment

Verify that the pods are created and runnig -

```bash
kubectl get pods -n tyk-sre
```
Verify that the service has been created -

```bash
kubectl get svc -n tyk-sre
```

## 6. Access locally

The application is exposed using a ClusterIP Service.

Use port-forwarding:

```bash
kubectl port-forward -n tyk-sre svc/tyk-sre-assignment-app 8080:8080
```

Then:

Test with the health endpoint -

```bash
curl http://localhost:8080/healthz
```

---

# API

## Health

```http
GET /healthz
```

Checks whether the application itself is healthy.

This endpoint does not depend on the Kubernetes API.

---

## Deployment Health

```http
GET /deployments/health
```

Returns health information for Kubernetes Deployments.

The health evaluation is based on Kubernetes Deployment status rather than simply checking whether a Deployment object exists.

---

## Kubernetes Connectivity

```http
GET /connectivity
```

Checks whether the application can communicate with the Kubernetes API server.

Example successful response:

```json
{
  "status": "ok",
  "message": "Successfully connected to Kubernetes API Server"
}
```

---

## Network Isolation

```http
POST /network/isolate
```

Creates the Kubernetes NetworkPolicy resources required for the requested isolation.

The request identifies workloads using Kubernetes namespaces and label selectors.

---

## Revert Network Isolation

```http
POST /network/revert
```

Removes the NetworkPolicies associated with the supplied isolation configuration.

Reverting an already-removed policy is treated as a successful no-op.

---

# Stories and Design Decisions

## Story 1 — Deployment Health

### Requirement

> As an SRE, I want to know whether all the deployments in the k8s cluster have as many healthy pods as request by the respective `Deployment` spec.

### Approach

The application queries Kubernetes Deployments using the Kubernetes Go client and evaluates the Deployment's observed state against its desired state.

The implementation considers Deployment rollout information rather than simply checking whether the Deployment object exists. The Deployment is considered healthy when the number of ready replicas is at least the desired replica count and the controller has observed the current Deployment generation.

A Deployment is considered healthy when:

- `status.readyReplicas >= spec.replicas`
- `status.observedGeneration >= metadata.generation`

The first check verifies that at least the desired number of replicas are ready.

The second check verifies that the Deployment controller has observed the latest Deployment specification.

### Design Decisions

#### Kubernetes is the source of truth

The application does not maintain its own Deployment state.

Kubernetes already maintains the authoritative state through the Deployment controller, so the tool reads that state rather than introducing another state-management mechanism.

#### Desired state vs observed state

The health check considers Kubernetes Deployment status rather than only checking resource existence.

Relevant status information includes:

- desired replicas
- updated replicas
- ready replicas
- available replicas
- observed generation
- Deployment conditions

This is important during rolling updates where old and new ReplicaSets can temporarily coexist.

#### Why not simply check `availableReplicas >= desiredReplicas`?

That can produce a false positive during a rollout.

For example:

```text
Desired replicas:     3
Updated replicas:    2
Available replicas:  5
```

There are enough available Pods to satisfy the desired count, but the latest revision has not yet reached the desired replica count.

Therefore, availability alone is not sufficient to determine rollout health.

### Trade-off

The implementation provides Deployment-level health rather than complete root-cause analysis.

An unhealthy Deployment may require further investigation into:

- Pod scheduling
- image pull failures
- container crashes
- resource constraints
- node health
- application-level health

Those diagnostics are intentionally outside the scope of this story.

---

## Story 2 — Network Isolation

### Requirement

> As an SRE, I want to prevent two workloads defined by k8s namespace(s) and label selectors from being able to exchange any network activity on demand.

### Approach

The application uses Kubernetes NetworkPolicy resources to implement workload isolation.

The isolation operation is:

```http
POST /network/isolate
```

and the reversion operation is:

```http
POST /network/revert
```

### Important NetworkPolicy Design Detail

Kubernetes NetworkPolicy is fundamentally an allow-based model.

A NetworkPolicy does not provide a traditional:

```text
DENY traffic from X
```

rule.

Instead, once traffic is isolated for a direction, traffic is allowed only when it matches an applicable allow rule. This might also impact ingress traffic from other workloads coming into the desired workload.

NetworkPolicies are also additive: multiple policies selecting the same Pods combine their allowed traffic.

This is an important constraint of the Kubernetes API and directly influenced the implementation.

### Why the implementation uses an ingress policy

The current implementation uses:

```text
policyTypes:
  - Ingress
```

The isolation is therefore expressed by controlling ingress to the selected destination workloads.

This means the implementation does not attempt to implement a custom deny rule that does not exist in the NetworkPolicy API.

The resulting behavior is dependent on the ingress policy semantics and the existing policies in the cluster.

### Design Decisions & Trade-Offs

#### Design Decision: Ingress-only vs Ingress + Egress

An alternative would have been to create both:

```text
Ingress
Egress
```

policies.

That would provide a more explicit two-direction isolation model, but it also changes the behavior and scope of the policy:

- more policies/rules to reason about
- greater potential to interfere with legitimate egress traffic
- more complex handling of existing policies
- broader blast radius

For the assignment, the implementation keeps the policy focused on ingress isolation.

This is a deliberate scope trade-off rather than an assumption that an ingress policy is equivalent to an explicit deny rule.

### Important Operational Consideration

Because NetworkPolicies are additive, the application cannot safely reason about isolation in complete isolation from policies that already exist in the cluster.

For example, another NetworkPolicy selecting the same destination Pods can contribute additional allowed ingress traffic.

Therefore, the implementation should be understood as applying Kubernetes NetworkPolicy semantics rather than implementing an absolute firewall-style deny rule.

The implementation intentionally does not attempt to:

- manipulate the CNI directly
- inspect or rewrite every existing NetworkPolicy
- implement custom deny semantics
- build a complete network policy management system

The trade-off is simplicity and Kubernetes-native behavior at the cost of not providing an absolute firewall-style isolation guarantee independent of other policies.

---

## Story 3 — Kubernetes API Connectivity

### Requirement

> As an SRE, I want to always know whether the application can connect to the Kubernetes API server.

### Approach

The application exposes:

```http
GET /connectivity
```

The handler uses the Kubernetes discovery client to verify communication with the Kubernetes API server.

### Design Decisions & Tradeoffs

#### Reuse the Kubernetes client configuration

The connectivity check uses the same Kubernetes client configuration used by the rest of the application.

This avoids introducing a separate connection mechanism and ensures that the connectivity check validates the same authentication/configuration used by the application.

#### Why use the Discovery API?

The application does not actually need the Kubernetes version for its business logic.

The discovery client's `ServerVersion()` call is used because it provides a lightweight authenticated request to the Kubernetes API server.

The important result is whether the request succeeds or fails.

#### Scope of the connecticity check

The endpoint checks connectivity rather than performing a full functional authorization test for every Kubernetes API operation.

A successful connectivity check therefore means:

> The application can communicate with the Kubernetes API server.

It does not necessarily mean:

> The application has permission to perform every Kubernetes operation required by every endpoint.

Authorization failures are surfaced by the individual operations that require those permissions.

---

## Story 4 — Container Image Build

### Requirement

> As an application developer, I want to build this application into a container image when I push a commit to the main branch of its repository.

### Approach

GitHub Actions performs:

```text
Push to main
     |
     v
+----------------------+
| Checkout repository  |
+----------+-----------+
           |
           v
+----------------------+
| Set up Go            |
| Version from go.mod  |
+----------+-----------+
           |
           v
+----------------------+
| Download & verify    |
| Go modules           |
+----------+-----------+
           |
           v
+----------------------+
| Go Quality Checks    |
|                      |
| - gofmt              |
| - go vet             |
| - go test            |
+----------+-----------+
           |
           v
+----------------------+
| Helm Validation      |
|                      |
| - helm lint          |
+----------+-----------+
           |
           v
+----------------------+
| Build Docker Image   |
|                      |
| :<commit-sha>        |
+----------+-----------+
           |
           v
+----------------------+
| Trivy Security Scan  |
|                      |
| OS packages          |
| Go/library deps      |
| All severities       |
+----------+-----------+
           |
           v
+----------------------+
| Generate Report      |
|                      |
| trivy-report.json    |
+----------+-----------+
           |
           +----------------------+
           |                      |
           v                      v
+----------------------+  +----------------------+
| GitHub Actions       |  | GitHub Actions       |
| Job Summary          |  | Artifact             |
|                      |  |                      |
| Vulnerability        |  | trivy-report-<sha>   |
| summary              |  |                      |
+----------------------+  +----------+-----------+
                                      |
                                      v
                           +----------------------+
                           | Push Docker Image    |
                           |                      |
                           | :<commit-sha>        |
                           | :latest              |
                           +----------+-----------+
                                      |
                                      v
                           +----------------------+
                           | Published Artifact   |
                           |                      |
                           | GHCR                 |
                           +----------------------+
```

### Design Decisions

#### Go version

The project uses the Go version declared by the project configuration.

The CI environment should remain compatible with the project's declared Go version rather than silently upgrading the compiler as part of the pipeline.

This reduces the risk of CI passing against an untested toolchain that differs from local development.

#### Multi-stage build

The Dockerfile separates the build environment from the runtime environment.

Build dependencies remain in the builder stage rather than being included in the final application image.

This keeps the runtime image smaller and reduces unnecessary packages in the production container.

#### Static Linux binary

The application is built with:

```text
CGO_ENABLED=0
GOOS=linux
```

This produces a Linux binary without CGO runtime dependencies.

### Container Security

The CI pipeline scans every built container image using Trivy before publication.

The scan covers:

- Operating-system packages
- Application/library dependencies
- Known vulnerabilities across all severities

The security workflow is intentionally designed as:

```text
Build image
    |
    v
Scan exact image with Trivy
    |
    v
Generate vulnerability report
    |
    v
Publish report as CI artifact
    |
    v
Push the scanned image
```

The scan is non-blocking and produces a JSON vulnerability report that is retained as a GitHub Actions artifact for 14 days. A summary is also published to the GitHub Actions workflow summary.

The scan is performed against the immutable commit-SHA image, ensuring the security report corresponds to the exact artifact being published. Vulnerabilities without an available fix are ignored to avoid reporting them as actionable findings. The current implementation treats Trivy as a visibility and security-awareness mechanism rather than a release gate. A production environment could introduce severity-based blocking policies based on organizational risk tolerance.

#### Commit SHA image tag

The image is tagged using the Git commit SHA:

```text
<image>:<commit-sha>
```

This provides a deterministic reference to the exact source revision used to build the image.

#### `latest` tag

The `latest` tag is also published as a convenience tag representing the most recent build.It is mutable and therefore should not be preferred when reproducibility is important.

### CI Trade-off

The pipeline is intentionally simple:

```text
test -> build -> publish
```

It does not attempt to implement a complete production release platform with environment promotion, signing, vulnerability scanning, or deployment automation because those were outside the scope of the assignment.

The important property is that an image is not published before the automated tests pass.

---

## Story 5 — Helm Deployment

### Requirement

> As an application developer, I want to be able to deploy this application into a Kubernetes cluster using Helm.

### Approach

The application is packaged as:

```text
helm/tyk-sre-assignment
```

The chart manages the Kubernetes resources required to run the application, including:

- Deployment
- Service
- ServiceAccount
- RBAC
- Health probes

### Design Decisions

#### Dedicated namespace

The application resources is deployed into a dedicated namespace rather than `default` which has the following -

- clearer resource ownership
- easier lifecycle management
- less chance of collision with unrelated workloads
- cleaner operational separation

The application may still require cluster-scoped RBAC because it operates on resources across namespaces.

#### Dedicated ServiceAccount

The application runs using its own ServiceAccount instead of the Kubernetes `default` ServiceAccount.

This provides an explicit identity for the application and allows RBAC permissions to be attached to that identity.

#### RBAC scope

The application can operate across namespaces, so the required permissions are represented using cluster-scoped RBAC where necessary.

The intent is to grant the permissions required by the tool rather than relying on broad administrative access.

#### Health probes

The application uses `/healthz` for:

- startup probe
- readiness probe
- liveness probe

The endpoint does not depend on the Kubernetes API.

This separates application health from dependency health.

The rationale is:

```text
/healthz
    |
    +-- Application health

/connectivity
    |
    +-- Kubernetes dependency health
```

A Kubernetes API outage should therefore be observable through `/connectivity` without automatically causing the application container to be restarted.

### Trade-off

The Helm chart intentionally provides the resources required by the assignment rather than exposing every possible Kubernetes configuration option.

This keeps the chart understandable and reduces configuration complexity while still allowing important deployment parameters such as the image repository and tag to be configured.

---

# Limitations

The implementation is intentionally scoped to the five assignment stories.

- Network isolation depends on the cluster's CNI supporting and enforcing NetworkPolicy.
- Kubernetes NetworkPolicy is an allow-based, additive model rather than an explicit deny-based firewall model.
- The current isolation implementation uses ingress policy semantics and therefore does not represent a general bidirectional firewall.
- Existing NetworkPolicies selecting the same workloads can affect the final effective network behavior.
- Deployment health provides Deployment-level health rather than complete root-cause diagnostics.
- The application is primarily exposed as an HTTP API rather than a dedicated CLI. We can build a cli wrapper using the exisiting functionality if that is required.
- The `latest` image tag is mutable; SHA-based tags are preferred for reproducible deployments.

---