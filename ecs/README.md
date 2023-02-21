# Running with Elastic Container Service

> https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_AWSCLI_Fargate.html
> https://docs.aws.amazon.com/cli/latest/userguide/cli-services-ec2-sg.html
> https://docs.aws.amazon.com/cli/latest/reference/ecs/create-service.html
> https://docs.aws.amazon.com/cli/latest/reference/elbv2/create-load-balancer.html
> https://docs.aws.amazon.com/cli/latest/reference/ecs/register-task-definition.html
> https://docs.aws.amazon.com/pt_br/elasticloadbalancing/latest/application/tutorial-application-load-balancer-cli.html

## Preparing resources

```bash
REGION="us-east-1"
API_NAME="go-micro-api"
# ecs config
ECS_CLUSTER="$API_NAME-cluster"
ECS_SG_NAME="$API_NAME-sg"
# load balancer config
ALB_NAME="$API_NAME-alb"
ALB_SG_NAME="$API_NAME-alb-sg"
ALB_TG_NAME="$API_NAME-tg"

# vpc config, create vpc with 2 public and 2 private subnets
VPC_ID="vpc-00000000000000000"
PUBLIC_SUBNET_IDS="subnet-00000000000000000,subnet-00000000000000000"
PRIVATE_SUBNET_IDS="subnet-00000000000000000,subnet-00000000000000000"

# create role for ecs AmazonECSTaskExecutionRole
aws iam create-role \
  --role-name AmazonECSTaskExecutionRole \
  --assume-role-policy-document file://./AmazonECSTaskExecutionRole.json

aws iam attach-role-policy \
  --policy-arn "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy" \
  --role-name AmazonECSTaskExecutionRole

aws iam attach-role-policy \
  --policy-arn "arn:aws:iam::aws:policy/service-role/AmazonECS_FullAccess" \
  --role-name AmazonECSTaskExecutionRole

# create ecs cluster
aws ecs create-cluster --region $REGION --cluster-name $ECS_CLUSTER

export ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text)
echo $ACCOUNT_ID

sed -i "s/{{ACCOUNT_ID}}/${ACCOUNT_ID}/; s/{{REGION}}/${REGION}/; s/{{API_NAME}}/${API_NAME}/" task.json

aws ecs register-task-definition --region $REGION --cli-input-json file://./task.json

TASK_DEF=$(aws ecs list-task-definitions --region $REGION --query 'taskDefinitionArns[0]' --output text | awk -F'/' '{print $2}')

# create SG for load balancer
ALB_SG_ID=$(aws ec2 create-security-group --region $REGION --group-name $ALB_SG_NAME --description $ALB_SG_NAME --vpc-id $VPC_ID --query 'GroupId' --output text)

aws ec2 authorize-security-group-ingress --region $REGION --group-id $ALB_SG_ID --protocol tcp --port 80 --cidr 0.0.0.0/0

# create SG for API
ECS_SG_ID=$(aws ec2 create-security-group --region $REGION --group-name $ECS_SG_NAME --description $ECS_SG_NAME --vpc-id $VPC_ID --query 'GroupId' --output text)

aws ec2 authorize-security-group-ingress --region $REGION --group-id $ECS_SG_ID --protocol tcp --port 9000 --source-group $ALB_SG_ID

# create target group
ALB_TG_ARN=$(aws elbv2 create-target-group \
  --region $REGION \
  --name $ALB_TG_NAME \
  --protocol HTTP \
  --port 80 \
  --health-check-port 9000 \
  --health-check-path "/api/v1/health/live" \
  --health-check-protocol HTTP \
  --target-type ip \
  --vpc-id $VPC_ID \
  --query 'TargetGroups[0].TargetGroupArn' --output text)

# create load balancer
ALB_ARN=$(aws elbv2 create-load-balancer \
  --region $REGION \
  --name $ALB_NAME \
  --scheme internet-facing \
  --subnets $(echo $PUBLIC_SUBNET_IDS | tr -s ',' ' ') \
  --security-groups $ALB_SG_ID \
  --query 'LoadBalancers[0].LoadBalancerArn' --output text)

# create listener
aws elbv2 create-listener --region $REGION \
  --load-balancer-arn $ALB_ARN \
  --protocol HTTP --port 80 \
  --default-actions Type=forward,TargetGroupArn=$ALB_TG_ARN

# create service bound to the target group
sed -i "s/{{ALB_TG_ARN}}/${ALB_TG_ARN}/; s/{{API_NAME}}/${API_NAME}/" service.json

aws ecs create-service --region $REGION \
  --cluster $ECS_CLUSTER \
  --service-name $API_NAME \
  --task-definition $TASK_DEF \
  --desired-count 1 \
  --launch-type "FARGATE" \
  --network-configuration "awsvpcConfiguration={subnets=[$PRIVATE_SUBNET_IDS],securityGroups=[$ECS_SG_ID],assignPublicIp=DISABLED}" \
  --cli-input-json file://./service.json

aws ecs list-services --region $REGION --cluster $ECS_CLUSTER

# check load balancer DNS
ALB_DNS=$(aws elbv2 describe-load-balancers --region $REGION --name $ALB_NAME --query 'LoadBalancers[0].DNSName' --output text)

curl --url "http://${ALB_DNS}/api/v1/message"
# {"data":"Hello World From ECS","statusCode":200}
```
