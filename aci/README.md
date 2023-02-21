# Running with Azure Container Instances

> https://learn.microsoft.com/pt-BR/azure/container-instances/container-instances-multi-container-yaml

> https://learn.microsoft.com/en-us/azure/container-instances/container-instances-quickstart

## Preparing resources

```bash
LOCATION="eastus"
RESOURCE_GROUP="go-micro-api-rg"
API_NAME="go-micro-api"
# log analytics
LOG_ANALYTICS_WORKSPACE_NAME="$API_NAME-log-analytics"
# network setting
VNET_NAME="$API_NAME-vnet"
SUBNETS_PREFIX_NAME="$API_NAME-subnet"
LB_IP_NAME="$API_NAME-pub-ip"


# create resource group
az group create --name $RESOURCE_GROUP --location $LOCATION


# network config
LB_SUBNET_NAME="${SUBNETS_PREFIX_NAME}-a"
az network vnet create \
  --name $VNET_NAME \
  --resource-group $RESOURCE_GROUP \
  --location $LOCATION \
  --address-prefix 10.100.0.0/20 \
  --subnet-name "$LB_SUBNET_NAME" \
  --subnet-prefix 10.100.0.0/24

CONTAINER_SUBNET_NAME="${SUBNETS_PREFIX_NAME}-b"
CONTAINER_SUBNET_ID=$(az network vnet subnet create \
  --name "$CONTAINER_SUBNET_NAME" \
  --resource-group $RESOURCE_GROUP \
  --vnet-name $VNET_NAME \
  --address-prefix 10.100.1.0/24 \
  --query "id" --out tsv)

# replace config on yaml file
sed -i "s/{{SUBNET_ID}}/${CONTAINER_SUBNET_ID}/" container-group.yaml
sed -i "s/{{SUBNET_NAME}}/${CONTAINER_SUBNET_NAME}/" container-group.yaml


# create log analytics workspace
az monitor log-analytics workspace create --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --location $LOCATION

WORKSPACE_ID=$(az monitor log-analytics workspace show --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --query "customerId" --out tsv)
# replace config on yaml file
sed -i "s/{{WORKSPACE_ID}}/${WORKSPACE_ID}/" container-group.yaml

WORKSPACE_KEY=$(az monitor log-analytics workspace get-shared-keys --resource-group $RESOURCE_GROUP --workspace-name $LOG_ANALYTICS_WORKSPACE_NAME --query "primarySharedKey" --out tsv)
WORKSPACE_KEY=$(echo "$WORKSPACE_KEY" | sed -r 's/\//\\\//gm')
# replace config on yaml file
sed -i "s/{{WORKSPACE_KEY}}/${WORKSPACE_KEY}/" container-group.yaml
```

## Creating the load balancer and container group

```bash
# create container group
az container create --resource-group $RESOURCE_GROUP --file container-group.yaml
# retrieve private container ip
CONTAINER_IP=$(az container show \
  --resource-group $RESOURCE_GROUP \
  --name $API_NAME \
  --query ipAddress.ip --output tsv)


# create application gateway
az network public-ip create \
  --resource-group $RESOURCE_GROUP \
  --name $LB_IP_NAME \
  --allocation-method Static \
  --sku Standard

az network application-gateway create \
  --name myAppGateway \
  --location $LOCATION \
  --resource-group $RESOURCE_GROUP \
  --capacity 2 \
  --sku Standard_v2 \
  --http-settings-protocol http \
  --public-ip-address $LB_IP_NAME \
  --vnet-name $VNET_NAME \
  --subnet $LB_SUBNET_NAME \
  --servers "$CONTAINER_IP"


# show public ip
LB_PUB_IP=$(az network public-ip show \
  --resource-group $RESOURCE_GROUP \
  --name $LB_IP_NAME \
  --query [ipAddress] --output tsv)

curl --url "http://${LB_PUB_IP}:9000/api/v1/message"
# {"data":"Hello World From ACI","statusCode":200}


# get logs from container
az container logs --resource-group $RESOURCE_GROUP --name $API_NAME --follow

# query to get some logs on log analytics
ContainerInstanceLog_CL
| project parse_json(Message)
| project
  Host = Message.host,
  Ip = Message.ip,
  Msg = Message.message,
  Method = Message.method,
  Path = Message.path,
  Severity = Message.severity,
  Timestamp = Message.timestamp
| where Path hasprefix "/api/v1/message"
| order by totimespan(Timestamp) desc
| take 10


# clean up
az group delete --name $RESOURCE_GROUP --yes
```
