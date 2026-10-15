# exist.io Enterprise Kubernetes Helm Chart

Official Helm chart for installing and managing high-availability `exist.io` self-hosted enterprise Git forge on Kubernetes.

## Prerequisites

- Kubernetes 1.25+
- Helm 3.10+
- Storage provisioner supporting `ReadWriteOnce` persistent volumes (AWS EBS, GCP Persistent Disk, Azure Disk, Ceph, Longhorn)
- Ingress controller (e.g., `ingress-nginx`)
- Optional: `cert-manager` for automated TLS certificates
- Optional: `prometheus-operator` for scraping `/metrics` via ServiceMonitor

## Installation

```bash
# Add local or remote chart repository
helm repo add exist https://charts.exist.io
helm repo update

# Install chart into 'exist' namespace
helm install exist-cluster ./deploy/helm/exist-enterprise \
  --namespace exist \
  --create-namespace \
  --set ingress.hosts[0].host=git.enterprise.internal \
  --set ingress.tls[0].hosts[0]=git.enterprise.internal \
  --set persistence.size=200Gi
```

## Production Architecture

1. **Storage Persistence**: Bare Git repositories and SQLite metadata WAL database persist on a high-IOPS persistent volume mounted at `/var/lib/exist`.
2. **Dual-Protocol Ingress**:
   - HTTP/HTTPS (Port 80/443 -> Container 8080): Web UI, REST API, Git Smart HTTP, and Git LFS.
   - Pure-Go SSH (Port 22/2222): Zero external dependencies, enrolled SSH public keys & certificate authorities.
3. **Observability**: Exposes real-time Prometheus metrics at `/metrics` scraped by Prometheus Operator `ServiceMonitor`.
