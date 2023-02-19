# Using Helm

```bash
RELEASE_NAME="http-simple-api"


helm create "http-simple-api"

mv "http-simple-api" helm
cd helm/


helm install "http-simple-api" ./ --debug --wait --timeout 15m

helm upgrade -i "http-simple-api" ./ --debug --wait --timeout 15m

helm ls

helm template "http-simple-api" --set image.tag="v1.0.0" ./ > deploy.yaml

helm rollback "http-simple-api" 1


helm test "http-simple-api" --debug --timeout 1m


helm delete "http-simple-api" --debug
```
