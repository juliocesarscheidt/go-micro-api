# Using Helm chart

```bash
RELEASE_NAME="http-simple-api"

# create chart
helm create "application"

mv "application" helm

# install and update
helm install "$RELEASE_NAME" ./helm --debug --wait --timeout 15m

helm upgrade -i "$RELEASE_NAME" ./helm --debug --wait --timeout 15m

# list releases
helm ls

# rollback a release
helm rollback "$RELEASE_NAME" 1

# generate template
helm template "$RELEASE_NAME" --set image.tag="v1.0.0" ./helm > deployment.yaml

# test connection
helm test "$RELEASE_NAME" --debug --timeout 1m

# delete a release
helm delete "$RELEASE_NAME" --debug
```
