# Simple Golang API

## Build

```bash
docker image build --tag juliocesarmidia/http-simple-api:v1.0.0 .

docker image ls \
  --format "table {{.ID}}\t{{.Repository}}\t{{.Tag}}\t{{.Size}}" --filter="reference=juliocesarmidia/http-simple-api:v1.0.0"

docker image push juliocesarmidia/http-simple-api:v1.0.0

docker image history juliocesarmidia/http-simple-api:v1.0.0 --no-trunc

docker container run -d \
  --name http-simple-api \
  --publish 9000:9000 \
  --cap-drop NET_BIND_SERVICE \
  --sysctl net.ipv4.ip_unprivileged_port_start=1024 \
  --memory='16MB' \
  --cpus='0.1' \
  --env MESSAGE="$(uptime -s)" \
  --restart on-failure \
  juliocesarmidia/http-simple-api:v1.0.0

docker container update --memory='8MB' http-simple-api

docker container stats http-simple-api --no-stream
docker container top http-simple-api

docker container inspect http-simple-api

docker container logs -f --tail 100 http-simple-api

curl --url 'http://localhost:9000/api/v1/message'
curl --url 'http://localhost:9000/api/v1/health/live'
curl --url 'http://localhost:9000/api/v1/health/ready'

docker container rm -f http-simple-api
```

## Running with Kubernetes

```bash
kubectl apply -f deployment.yaml

INGRESS_IP=$(kubectl get service -n ingress-nginx -l app.kubernetes.io/instance=ingress-nginx --no-headers | tr -s ' ' ' ' | cut -d ' ' -f 3)
echo "${INGRESS_IP} api.golang.local" >> /etc/hosts

curl --url 'http://api.golang.local/api/v1/message'
curl --url 'http://api.golang.local/api/v1/health/live'
curl --url 'http://api.golang.local/api/v1/health/ready'

kubectl get pod,svc,ingress -n default

kubectl logs -f -l component=api -n default --tail 100

echo "127.0.0.1 api.golang.local" >> /etc/hosts
curl -H 'Host: api.golang.local' 'http://127.0.0.1/api/v1'

kubectl delete -f deployment.yaml
```
