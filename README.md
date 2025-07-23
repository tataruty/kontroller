# Kontroller

uses config files from <https://github.com/tataruty/ngf_test_app/tree/main> repo

## To build

build image:

`docker build . -t kontroller:latest --platform linux/amd64,linux/arm64`

add tag:

`docker tag kontroller:latest tusova194/kontroller`

push:

`docker push tusova194/kontroller`

## To run

1. add gateway CRDs:

`kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml`
2. run helm installation:

`helm install kontroller-release ./charts/kontroller --set fullnameOverride=full-overrrmi`

## To uninstall

`helm uninstall kontroller-release`

## To test

### Service

conntect to logs:

`kubectl logs -f {{@pod_name}}`

create namespace:

`kubectl create namespace TNS`

deploy some service:

```yaml
kubectl apply -f - -n TNS <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-test-app
  template:
    metadata:
      labels:
        app: my-test-app
    spec:
      containers:
      - name: my-test-app
        image: tusova194/my_test_app:1.0.5
        ports:
        - containerPort: 3002

---
apiVersion: v1
kind: Service
metadata:
  name: my-app-service
spec:
  ports:
  - port: 80
    targetPort: 3002
    protocol: TCP
    name: http
  selector:
    app: my-test-app
EOF
```

check:

`kubectl get all -o wide -n TNS`

delete:

`deployment.apps/my-app -n TNS`

### HTTPRoute

create httpRoute:

```yaml
kubectl apply -f - -n TNS <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-app-route
spec:
  parentRefs:
  - name: gateway
    sectionName: http
  hostnames:
  - "test.my-apps.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /test
    backendRefs:
    - name: my-app-service
      port: 80
EOF
```

check:

`kubectl get httproutes -n TNS`

delete:

`kubectl delete httproute my-app-route -n TNS`
