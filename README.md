# Kubernetes Demo Application (kuard)

A simple Go application that demonstrates how Kubernetes works, particularly focusing on liveness and readiness probes.

## Features

- Display request information
- Display server information
- Kubernetes liveness probe endpoint
- Kubernetes readiness probe endpoint
- Easy toggle for enabling/disabling probes

## Endpoints

- `/` - Home page with links to all endpoints
- `/request` - Display detailed information about the HTTP request
- `/server` - Display information about the server (hostname, OS, Go version, etc.)
- `/healthz` - Kubernetes liveness probe endpoint
- `/readyz` - Kubernetes readiness probe endpoint
- `/toggle/ready` - Toggle the readiness status (for demonstration purposes)
- `/toggle/healthy` - Toggle the health status (for demonstration purposes)

## Building and Running

### Local Development

```bash
# Build the application
go build -o kuard .

# Run the application
./kuard
```

The application will be available at http://localhost:8080

### Docker

```bash
# Build the Docker image
docker build -t kuard:latest .

# Run the Docker container
docker run -p 8080:8080 kuard:latest
```

The application will be available at http://localhost:8080

## Kubernetes Deployment

Create a Kubernetes deployment file (kuard-deployment.yaml):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kuard
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kuard
  template:
    metadata:
      labels:
        app: kuard
    spec:
      containers:
      - name: kuard
        image: kuard:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 3
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 3
```

Create a Kubernetes service file (kuard-service.yaml):

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kuard
spec:
  selector:
    app: kuard
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy to Kubernetes:

```bash
# Apply the deployment
kubectl apply -f kuard-deployment.yaml

# Apply the service
kubectl apply -f kuard-service.yaml
```

## Testing Kubernetes Probes

To test the readiness probe:
1. Access the application
2. Navigate to `/toggle/ready` to toggle the readiness status
3. Observe in Kubernetes that the pod is removed from service when not ready

To test the liveness probe:
1. Access the application
2. Navigate to `/toggle/healthy` to toggle the health status
3. Observe in Kubernetes that the pod is restarted when not healthy

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) to learn how you can help.

## Code of Conduct

This project adopts the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Security

For information about reporting security issues, please see our [Security Policy](SECURITY.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.