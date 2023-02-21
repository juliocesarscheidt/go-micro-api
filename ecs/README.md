
```bash

ECS_REGION="us-east-1"
ECS_CLUSTER="go-micro-api-cluster"

aws ecs create-cluster --region $ECS_REGION --cluster-name $ECS_CLUSTER


export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo $ACCOUNT_ID

sed -i "s/{{ACCOUNT_ID}}/${ACCOUNT_ID}/" task.json
sed -i "s/{{ECS_REGION}}/${ECS_REGION}/" task.json

aws ecs register-task-definition --region $ECS_REGION --cli-input-json file://./task.json


aws ecs create-service


```
