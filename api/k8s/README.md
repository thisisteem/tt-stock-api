# TT Stock API - Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the TT Stock API to a Kubernetes cluster.

## 📁 Files Overview

- `namespace.yaml` - Creates the team-dev namespace
- `configmap.yaml` - Application configuration
- `secret.yaml` - Sensitive data (passwords, JWT secrets)
- `persistent-volumes.yaml` - Storage for PostgreSQL and Redis
- `migrations-configmap.yaml` - Database migration files
- `postgres-deployment.yaml` - PostgreSQL database deployment
- `redis-deployment.yaml` - Redis cache deployment
- `api-deployment.yaml` - TT Stock API application deployment
- `ingress.yaml` - External access configuration

## 🚀 Quick Deployment

### 1. Build and Push Docker Image

```bash
# Build production image
docker build -t ttstock-api:latest .

# Tag for your registry
docker tag ttstock-api:latest thisisteem/ttstock-api:latest

# Push to registry
docker push thisisteem/ttstock-api:latest
```

### 2. Update Image in Deployment

Edit `api-deployment.yaml` and change:
```yaml
image: ttstock-api:latest  # Change this to your registry
```

### 3. Deploy to Kubernetes

```bash
# Deploy individually
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f persistent-volumes.yaml
kubectl apply -f migrations-configmap.yaml
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml
kubectl apply -f api-deployment.yaml
kubectl apply -f ingress.yaml
```

### 4. Check Deployment Status

```bash
# Check pods
kubectl get pods -n team-dev

# Check services
kubectl get services -n team-dev

# Check ingress
kubectl get ingress -n team-dev

# View logs
kubectl logs -f deployment/tt-stock-api -n team-dev
```

## 🔧 Configuration

### Environment Variables

The application uses ConfigMaps and Secrets for configuration:

**ConfigMap** (`configmap.yaml`):
- Server settings
- Database connection (non-sensitive)
- JWT settings (non-sensitive)
- Application settings

**Secret** (`secret.yaml`):
- Database password
- JWT secret key

### Security Notes

⚠️ **IMPORTANT**: Change the default secrets in `secret.yaml`:

```bash
# Generate secure passwords
echo -n "your-secure-db-password" | base64
echo -n "your-secure-jwt-secret-key" | base64
```

Update the `data` section in `secret.yaml` with your encoded values.

### Storage

- **PostgreSQL**: 1Gi persistent volume
- **Redis**: Stateless (no persistent storage)

Adjust storage sizes in `persistent-volumes.yaml` based on your needs.

### Ingress

The ingress is configured for `tt-stock-api.local`. Update the host in `ingress.yaml`:

```yaml
rules:
- host: your-domain.com  # Change this
```

## 📊 Scaling

### Horizontal Pod Autoscaling

Add HPA for automatic scaling:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: tt-stock-api-hpa
  namespace: team-dev
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: tt-stock-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Scaling

Adjust resource limits in `api-deployment.yaml`:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

## 🔍 Monitoring

### Health Checks

The deployment includes:
- **Liveness Probe**: `/health` endpoint
- **Readiness Probe**: `/health` endpoint

### Logs

```bash
# View API logs
kubectl logs -f deployment/tt-stock-api -n team-dev

# View PostgreSQL logs
kubectl logs -f deployment/postgres -n team-dev

# View Redis logs
kubectl logs -f deployment/redis -n team-dev
```

## 🧹 Cleanup

```bash
# Delete individually
kubectl delete -f ingress.yaml
kubectl delete -f api-deployment.yaml
kubectl delete -f redis-deployment.yaml
kubectl delete -f postgres-deployment.yaml
kubectl delete -f migrations-configmap.yaml
kubectl delete -f persistent-volumes.yaml
kubectl delete -f secret.yaml
kubectl delete -f configmap.yaml
kubectl delete -f namespace.yaml

# Or delete everything at once
kubectl delete namespace team-dev
```

## 🔗 Access Points

After deployment:

- **API**: `http://tt-stock-api.local` (or your configured domain)
- **Health Check**: `http://tt-stock-api.local/health`
- **Database**: Internal service `postgres-service:5432`
- **Redis**: Internal service `redis-service:6379`

## 📝 Notes

- The deployment uses 1 replica of the API
- PostgreSQL is single-instance with 1Gi persistent storage
- Redis is stateless (no persistent storage)
- All services are in the `team-dev` namespace
- PostgreSQL data survives pod restarts
- Ingress provides external access to the API
