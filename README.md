# Simple Golang API

## Build

```bash
docker image build --tag juliocesarmidia/http-simple-api:v1.0.0 .

docker image build --build-arg API_PORT=9000 --tag juliocesarmidia/http-simple-api:v1.0.0 .

docker image ls \
  --format "table {{.ID}}\t{{.Repository}}\t{{.Tag}}\t{{.Size}}" --filter="reference=juliocesarmidia/http-simple-api:v1.0.0"

docker image push juliocesarmidia/http-simple-api:v1.0.0

docker image history juliocesarmidia/http-simple-api:v1.0.0 --no-trunc

PID_LIMIT=$(cat /proc/sys/kernel/pid_max) # 131072
# 64-bit systems (up to 4.194.304)
echo 4194304 > /proc/sys/kernel/pid_max

docker container run -d \
  --name http-simple-api \
  --publish 9000:9000 \
  --cap-drop NET_BIND_SERVICE \
  --sysctl net.ipv4.ip_unprivileged_port_start=1024 \
  --memory='16MB' \
  --cpus='0.1' \
  --pids-limit $PID_LIMIT \
  --env MESSAGE="$(uptime -s)" \
  --env API_PORT=9000 \
  --restart on-failure \
  juliocesarmidia/http-simple-api:v1.0.0

docker container update --memory='8MB' http-simple-api

docker container stats http-simple-api --no-stream
docker container top http-simple-api

docker container inspect http-simple-api

docker container logs -f --tail 100 http-simple-api

curl localhost:9000/api/v1/
curl localhost:9000/api/v1/healthcheck

docker container rm -f http-simple-api
```

## Running with Kubernetes

```bash
kubectl apply -f api.yaml

INGRESS_IP=$(kubectl get service -n ingress-nginx -l app.kubernetes.io/instance=ingress-nginx --no-headers | tr -s ' ' ' ' | cut -d ' ' -f 3)
echo "${INGRESS_IP} api.blackdevs.local" >> /etc/hosts

curl 'http://api.blackdevs.local/api/v1/'
curl 'http://api.blackdevs.local/api/v1/healthcheck'


kubectl get pod,svc,ingress -n blackdevs

kubectl logs -f -l component=api -n blackdevs --tail 100

echo "127.0.0.1 api.blackdevs.local" >> /etc/hosts
curl -H 'Host: api.blackdevs.local' 'http://127.0.0.1/api/v1'

kubectl delete -f api.yaml
```
