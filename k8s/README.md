# HSA Receipt Manager - Kubernetes Deployment

Complete Kubernetes deployment configuration for the HSA Receipt Manager application stack on MicroK8s/K3s.

## Architecture Overview

The application consists of four main components:

```
┌─────────────────────────────────────────────────────────────┐
│                        hsa-app Namespace                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Frontend   │───▶│   API (Go)   │───▶│ OCR Service  │  │
│  │  (Vue/Nginx) │    │   Port 8080  │    │  (FastAPI)   │  │
│  │  NodePort    │    │   NodePort   │    │  Port 8000   │  │
│  │    30591     │    │    30081     │    │  ClusterIP   │  │
│  └──────────────┘    └──────┬───────┘    └──────────────┘  │
│                              │                                │
│                              │                                │
│                              ▼                                │
│                    ┌──────────────────┐                      │
│                    │  Shared Storage  │                      │
│                    │  (PVC/PV)        │                      │
│                    │  /data/hsa       │                      │
│                    └──────────────────┘                      │
└───────────────────────────┬───────────────────────────────┘
                             │
                             │ Cross-namespace access
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                     postgres-db Namespace                     │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│                    ┌──────────────────┐                       │
│                    │    PostgreSQL    │                       │
│                    │    Port 5432     │                       │
│                    │  Database: hsa_app│                      │
│                    │  User: hsa        │                       │
│                    └──────────────────┘                       │
└──────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Kubernetes cluster (MicroK8s, K3s, or standard K8s)
- `kubectl` configured to access your cluster
- Existing PostgreSQL deployment in `postgres-db` namespace
- Container registry accessible from cluster
- Sufficient storage for receipt images (~10GB recommended)

## File Structure

```
k8s/
├── namespace.yaml                    # Creates hsa-app namespace
├── pvc.yaml                          # Persistent volume for receipt storage
├── hsa-app-secret.yaml.example      # Template for database credentials
├── hsa-app-network-policy.yaml      # Allow access to postgres-db namespace
├── ocr-deployment.yaml              # OCR service with Claude API
├── api-deployment.yaml              # Go backend API
├── frontend-deployment.yaml         # Vue.js frontend
└── README.md                        # This file
```

## Quick Start

### 1. Prepare Your Environment

```bash
# Clone/navigate to k8s directory
cd k8s

# Update container image registry in all deployment files
# Replace "localhost:32000" with your registry
# Example: docker.io/yourusername or ghcr.io/yourusername
```

### 2. Create Database and User

Connect to your existing PostgreSQL instance:

```bash
# Get postgres pod name
kubectl get pods -n postgres-db

# Connect to PostgreSQL
kubectl exec -it <postgres-pod-name> -n postgres-db -- psql -U <admin-user> -d postgres

# Run these SQL commands:
CREATE DATABASE hsa_app;
CREATE USER hsa WITH PASSWORD 'your-secure-password-here';
GRANT ALL PRIVILEGES ON DATABASE hsa_app TO hsa;

# Connect to new database
\c hsa_app

# Grant schema permissions
GRANT ALL ON SCHEMA public TO hsa;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO hsa;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO hsa;

# Exit
\q
```

### 3. Configure Secrets

```bash
# Copy example secret file
cp hsa-app-secret.yaml.example hsa-postgres-secret.yaml

# Edit with your actual credentials
nano hsa-postgres-secret.yaml

# Update these values:
# - password: your-secure-password-here
# - connection_string: postgres://hsa:your-secure-password-here@...

# IMPORTANT: Add to .gitignore
echo "hsa-postgres-secret.yaml" >> ../.gitignore
```

### 4. Update Claude API Key

```bash
# Edit ocr-deployment.yaml
nano ocr-deployment.yaml

# Replace "YOUR-CLAUDE-API-KEY" with your actual Anthropic API key
```

### 5. Deploy the Stack

```bash
# 1. Create namespace
kubectl apply -f namespace.yaml

# 2. Create secrets
kubectl apply -f hsa-postgres-secret.yaml

# 3. Create network policy (allow access to postgres)
kubectl apply -f hsa-app-network-policy.yaml

# 4. Create storage
kubectl apply -f pvc.yaml

# 5. Deploy OCR service
kubectl apply -f ocr-deployment.yaml

# 6. Deploy API
kubectl apply -f api-deployment.yaml

# 7. Deploy frontend
kubectl apply -f frontend-deployment.yaml
```

### 6. Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n hsa-app

# Expected output:
# NAME                           READY   STATUS    RESTARTS   AGE
# api-go-xxxxxxxxxx-xxxxx        1/1     Running   0          2m
# ocr-service-xxxxxxxxxx-xxxxx   1/1     Running   0          3m
# frontend-xxxxxxxxxx-xxxxx      1/1     Running   0          1m

# Check services
kubectl get svc -n hsa-app

# Check logs
kubectl logs -l app=api-go -n hsa-app
kubectl logs -l app=ocr-service -n hsa-app
kubectl logs -l app=frontend -n hsa-app
```

## Access the Application

Once deployed, access the application via NodePort:

- **Frontend**: `http://<NODE-IP>:30591`
- **API**: `http://<NODE-IP>:30081/api/health`

To find your node IP:

```bash
kubectl get nodes -o wide
```

## Component Details

### Frontend (frontend-deployment.yaml)

**Purpose**: Vue.js SPA served by Nginx

**Configuration**:
- Port: 80 (internal), 30591 (NodePort)
- Image: `localhost:32000/hsa-app-frontend:latest`
- Resources: 128Mi memory, 50m CPU (request)

**Environment**: No environment variables needed (API URL auto-detected)

### API (api-deployment.yaml)

**Purpose**: Go backend REST API

**Configuration**:
- Port: 8080 (internal), 30081 (NodePort)
- Image: `localhost:32000/hsa-app-api:latest`
- Resources: 256Mi memory, 100m CPU (request)

**Environment Variables**:
- `DATABASE_URL`: PostgreSQL connection (from secret)
- `OCR_SERVICE_URL`: `http://ocr-service:8000`
- `HSA_DIR`: `/data/hsa`
- `PORT`: `8080`

**Volumes**: Mounts `hsa-storage` PVC at `/data/hsa`

### OCR Service (ocr-deployment.yaml)

**Purpose**: Claude AI-powered receipt OCR processing

**Configuration**:
- Port: 8000 (ClusterIP only, internal)
- Image: `localhost:32000/hsa-app-ocr:latest`
- Resources: 256Mi memory, 100m CPU (request)

**Environment Variables**:
- `CLAUDE_API_KEY`: Anthropic API key (from secret)
- `CLAUDE_MODEL`: `claude-3-5-haiku-20241022`
- `PYTHONUNBUFFERED`: `1`

**Security**: API key stored in Kubernetes secret

### Persistent Storage (pvc.yaml)

**Purpose**: Store receipt images

**Configuration**:
- Storage Class: `local-path`
- Access Mode: `ReadWriteMany`
- Size: 10Gi
- Mount Path: `/data/hsa`

**Directory Structure**:
```
/data/hsa/
├── unused/
│   ├── 2024/
│   └── 2025/
└── used/
    ├── 2024/
    └── 2025/
```

### Network Policy (hsa-app-network-policy.yaml)

**Purpose**: Allow hsa-app namespace to access PostgreSQL in postgres-db namespace

**Rules**:
- Allows ingress to postgres pod from hsa-app namespace
- Port: 5432 (TCP)
- Applied to: `postgres-db` namespace

## Database Schema

The application expects this schema in the `hsa_app` database:

```sql
CREATE TABLE IF NOT EXISTS receipts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    vendor VARCHAR(255),
    total_amount DECIMAL(10,2) NOT NULL,
    date DATE,
    hsa_qualified BOOLEAN DEFAULT true,
    hsa_status VARCHAR(20) DEFAULT 'Yes',
    image_path TEXT,
    image_hash VARCHAR(64),
    raw_text TEXT,
    used BOOLEAN DEFAULT false,
    used_date TIMESTAMP,
    use_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_receipts_user_id ON receipts(user_id);
CREATE INDEX idx_receipts_used ON receipts(used);
CREATE INDEX idx_receipts_image_hash ON receipts(image_hash);
CREATE INDEX idx_receipts_composite ON receipts(vendor, total_amount, date);
CREATE INDEX idx_receipts_hsa_status ON receipts(hsa_status);
```

**Note**: Migrations run automatically when the API starts.

## Configuration Checklist

Before deploying, ensure you've:

- [ ] Created `hsa_app` database and `hsa` user in PostgreSQL
- [ ] Updated `hsa-postgres-secret.yaml` with actual credentials
- [ ] Updated Claude API key in `ocr-deployment.yaml`
- [ ] Changed container image registry in all deployments
- [ ] Created storage directory on host (if using hostPath)
- [ ] Applied network policy for cross-namespace access
- [ ] Added `hsa-postgres-secret.yaml` to `.gitignore`

## Updating the Application

### Update a Single Component

```bash
# Rebuild and push image
docker build -t your-registry/hsa-app-api:latest ./api-go
docker push your-registry/hsa-app-api:latest

# Restart deployment
kubectl rollout restart deployment/api-go -n hsa-app

# Check rollout status
kubectl rollout status deployment/api-go -n hsa-app
```

### Update All Components

```bash
# Build and push all images
docker build -t your-registry/hsa-app-api:latest ./api-go
docker build -t your-registry/hsa-app-ocr:latest ./service-ocr
docker build -t your-registry/hsa-app-frontend:latest ./frontend

docker push your-registry/hsa-app-api:latest
docker push your-registry/hsa-app-ocr:latest
docker push your-registry/hsa-app-frontend:latest

# Restart all deployments
kubectl rollout restart deployment/api-go -n hsa-app
kubectl rollout restart deployment/ocr-service -n hsa-app
kubectl rollout restart deployment/frontend -n hsa-app
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n hsa-app

# Describe pod for events
kubectl describe pod <pod-name> -n hsa-app

# Check logs
kubectl logs <pod-name> -n hsa-app

# Check previous logs if crashed
kubectl logs <pod-name> -n hsa-app --previous
```

### Database Connection Issues

```bash
# Test connection from API pod
kubectl exec -it deployment/api-go -n hsa-app -- sh

# Or test from a temporary pod
kubectl run -it --rm debug --image=postgres:14-alpine -n hsa-app -- \
  psql postgres://hsa:your-password@postgres.postgres-db.svc.cluster.local:5432/hsa_app

# Inside psql:
\dt                    # List tables
SELECT * FROM receipts; # Query receipts
\q                     # Quit
```

### Storage Issues

```bash
# Check PVC status
kubectl get pvc -n hsa-app

# Check PV status
kubectl get pv

# Describe PVC for events
kubectl describe pvc hsa-storage -n hsa-app

# Check mounted storage in pod
kubectl exec -it deployment/api-go -n hsa-app -- ls -la /data/hsa
```

### Network Policy Issues

```bash
# Verify network policy exists
kubectl get networkpolicy -n postgres-db

# Test connectivity from API pod to PostgreSQL
kubectl exec -it deployment/api-go -n hsa-app -- \
  nc -zv postgres.postgres-db.svc.cluster.local 5432
```

### OCR Service Issues

```bash
# Check OCR service logs
kubectl logs -l app=ocr-service -n hsa-app

# Test OCR service from API pod
kubectl exec -it deployment/api-go -n hsa-app -- \
  wget -O- http://ocr-service:8000/health
```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `ImagePullBackOff` | Can't pull container image | Check registry URL and credentials |
| `CrashLoopBackOff` | Container keeps crashing | Check logs with `kubectl logs` |
| `Pending` | No resources or PVC issues | Check PVC status and node resources |
| `Connection refused` (postgres) | Network policy or wrong host | Verify network policy and connection string |
| `CLAUDE_API_KEY not configured` | Missing API key | Update secret in ocr-deployment.yaml |

## Monitoring

### Check Resource Usage

```bash
# Pod resource usage
kubectl top pods -n hsa-app

# Node resource usage
kubectl top nodes
```

### View Events

```bash
# Recent events in namespace
kubectl get events -n hsa-app --sort-by='.lastTimestamp'

# Watch events in real-time
kubectl get events -n hsa-app --watch
```

### Logs

```bash
# Follow logs in real-time
kubectl logs -f deployment/api-go -n hsa-app

# Get logs from all replicas
kubectl logs -l app=api-go -n hsa-app

# Export logs to file
kubectl logs deployment/api-go -n hsa-app > api-logs.txt
```

## Scaling

### Scale Deployments

```bash
# Scale API to 2 replicas
kubectl scale deployment/api-go --replicas=2 -n hsa-app

# Scale OCR service to 3 replicas
kubectl scale deployment/ocr-service --replicas=3 -n hsa-app

# Note: Frontend can be scaled freely
# API and OCR can be scaled, but ensure storage supports ReadWriteMany
```

### Autoscaling (Optional)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-go-hpa
  namespace: hsa-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-go
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Backup and Recovery

### Backup Database

```bash
# Export database
kubectl exec -it <postgres-pod> -n postgres-db -- \
  pg_dump -U hsa hsa_app > hsa_app_backup.sql

# Or use pg_dump from local machine
pg_dump -h <node-ip> -p <postgres-nodeport> -U hsa hsa_app > backup.sql
```

### Backup Receipt Images

```bash
# From the host machine where PV is mounted
tar -czf hsa_receipts_backup.tar.gz /data/hsa_receipts/

# Or copy from pod
kubectl cp hsa-app/api-go-xxx:/data/hsa ./local_backup/
```

### Restore Database

```bash
# Restore from backup
kubectl exec -i <postgres-pod> -n postgres-db -- \
  psql -U hsa hsa_app < hsa_app_backup.sql
```

## Security Best Practices

1. **Never commit secrets**: Add `hsa-postgres-secret.yaml` to `.gitignore`
2. **Use strong passwords**: Generate secure database passwords
3. **Rotate API keys**: Periodically update Claude API key
4. **Limit permissions**: Use network policies to restrict access
5. **Enable TLS**: Add Ingress with TLS for production
6. **Regular updates**: Keep container images updated
7. **Resource limits**: Set appropriate resource limits to prevent DoS
8. **RBAC**: Use Role-Based Access Control for kubectl access

## Production Recommendations

### Use Ingress Instead of NodePort

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hsa-ingress
  namespace: hsa-app
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - hsa.yourdomain.com
    secretName: hsa-tls
  rules:
  - host: hsa.yourdomain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-go
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend
            port:
              number: 80
```

### Use External Database

For production, consider managed PostgreSQL:
- Cloud providers (AWS RDS, Google Cloud SQL, Azure Database)
- Self-hosted with proper backups and replication

### Add Monitoring

Integrate with Prometheus and Grafana:
- Monitor pod health and resource usage
- Track API response times
- Alert on errors and high resource usage

### Implement Backup Automation

Create a CronJob for automated backups:
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-hsa
  namespace: hsa-app
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: your-backup-image
            # Add backup logic
```

## Cleanup

### Remove Application

```bash
# Delete all resources
kubectl delete -f frontend-deployment.yaml
kubectl delete -f api-deployment.yaml
kubectl delete -f ocr-deployment.yaml
kubectl delete -f pvc.yaml
kubectl delete -f hsa-app-network-policy.yaml
kubectl delete -f hsa-postgres-secret.yaml

# Optional: Delete namespace (will delete everything)
kubectl delete namespace hsa-app
```

### Remove Database

```bash
# Connect to PostgreSQL
kubectl exec -it <postgres-pod> -n postgres-db -- psql -U <admin> -d postgres

# Drop database and user
DROP DATABASE hsa_app;
DROP USER hsa;
\q
```

## Support

For issues or questions:
- Check logs: `kubectl logs -l app=<component> -n hsa-app`
- Review events: `kubectl get events -n hsa-app`
- Verify secrets: `kubectl get secrets -n hsa-app`
- Test connectivity: Use debug pods to test network access
