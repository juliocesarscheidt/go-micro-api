# Running with Azure Container Instances

> https://learn.microsoft.com/pt-BR/azure/container-instances/container-instances-multi-container-yaml

> https://learn.microsoft.com/en-us/azure/container-instances/container-instances-quickstart

## Preparing resources

```bash
RESOURCE_GROUP="go-micro-api-rg"
LOG_ANALYTICS_WORKSPACE_NAME="go-micro-api-log"
CONTAINER_NAME="go-micro-api"

# create resource group
az group create --name $RESOURCE_GROUP --location eastus

# create log analytics workspace
az monitor log-analytics workspace create --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --location eastus

WORKSPACE_ID=$(az monitor log-analytics workspace show --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --query "customerId" --out tsv)
sed -i "s/{{WORKSPACE_ID}}/${WORKSPACE_ID}/" container-group.yaml

WORKSPACE_KEY=$(az monitor log-analytics workspace get-shared-keys --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --query "primarySharedKey" --out tsv)
sed -i "s/{{WORKSPACE_KEY}}/${WORKSPACE_KEY}/" container-group.yaml
```

## Creating the container group

```bash
# create container group
az container create --resource-group $RESOURCE_GROUP --file container-group.yaml

API_HOSTNAME=$(az container show --resource-group $RESOURCE_GROUP --name $CONTAINER_NAME --query "ipAddress.fqdn" --out tsv)

curl --url "http://${API_HOSTNAME}:9000/api/v1/message"
# {"data":"Hello World From ACI","statusCode":200}

# get logs from container
az container logs --resource-group $RESOURCE_GROUP --name $CONTAINER_NAME --follow

# query to get some logs on log analytics
ContainerInstanceLog_CL
| project parse_json(Message)
| project Message.host, Message.message, Message.method, Message.path, Message.severity, Message.timestamp
| where Message_path == "/api/v1/message"

# clean up
az container delete --resource-group $RESOURCE_GROUP --name $CONTAINER_NAME

az monitor log-analytics workspace delete --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME

az group delete --name $RESOURCE_GROUP
```
