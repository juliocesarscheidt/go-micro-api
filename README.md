# Micro Golang API

Micro API made with Golang to run on containerized environments

## Using Makefile

```bash
make
```
![image](./images/make.PNG)

## Running with Docker

```bash
docker image build --tag juliocesarmidia/go-micro-api:v1.0.0 .

docker image ls \
  --format "table {{.ID}}\t{{.Repository}}\t{{.Tag}}\t{{.Size}}" \
  --filter="reference=juliocesarmidia/go-micro-api:v1.0.0"

docker image push juliocesarmidia/go-micro-api:v1.0.0

docker image history juliocesarmidia/go-micro-api:v1.0.0 --no-trunc

docker container run -d \
  --name go-micro-api \
  --publish 9000:9000 \
  --cap-drop NET_BIND_SERVICE \
  --sysctl net.ipv4.ip_unprivileged_port_start=1024 \
  --memory='16MB' \
  --cpus='0.1' \
  --env MESSAGE="$(uptime -s)" \
  --restart on-failure \
  juliocesarmidia/go-micro-api:v1.0.0

docker container update --memory='8MB' go-micro-api

docker container stats go-micro-api --no-stream
docker container top go-micro-api

docker container inspect go-micro-api

docker container logs -f --tail 100 go-micro-api

curl --url 'http://localhost:9000/api/v1/message'
curl --url 'http://localhost:9000/api/v1/health/live'
curl --url 'http://localhost:9000/api/v1/health/ready'

docker container rm -f go-micro-api
```

## Running with Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml

INGRESS_IP=$(kubectl get service -n ingress-nginx \
  -l app.kubernetes.io/instance=ingress-nginx --no-headers \
  | tr -s ' ' ' ' | cut -d' ' -f3)
echo "${INGRESS_IP} api.golang.local" >> /etc/hosts

curl --url 'http://api.golang.local/api/v1/message'
curl --url 'http://api.golang.local/api/v1/health/live'
curl --url 'http://api.golang.local/api/v1/health/ready'

kubectl get pod,svc,rs,ingress -n default

kubectl logs -f \
  -l app.kubernetes.io/name=go-micro-api \
  -n default --tail 100 --timestamps

kubectl delete -f k8s/deployment.yaml
```

## Testing API benchmark with siege

```bash
siege --time 30S --concurrent 100 \
  --benchmark 'http://localhost:9000/api/v1/message'
```
