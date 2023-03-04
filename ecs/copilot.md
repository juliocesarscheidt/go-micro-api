# AWS Copilot CLI

Copilot CLI uses CloudFormation behind to create the infrastructure

[https://docs.aws.amazon.com/AmazonECS/latest/developerguide/getting-started-aws-copilot-cli.html](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/getting-started-aws-copilot-cli.html)

> Install

```bash
curl -Lo /usr/local/bin/copilot --url 'https://github.com/aws/copilot-cli/releases/latest/download/copilot-linux' \
  && chmod +x /usr/local/bin/copilot
```

> Usage

```bash
# guided init
copilot init
# one-line init and deploy
copilot init \
  --app ecs-cluster \
  --name api \
  --type 'Load Balanced Web Service' \
  --dockerfile './src/Dockerfile' \
  --port 9000 \
  --deploy
# type 'Load Balanced Web Service'
# type 'Backend Service'

copilot env ls
copilot env show test

copilot app ls
copilot app show ecs-cluster

copilot svc ls
copilot svc show go-micro-api
copilot svc deploy
copilot svc status
copilot svc logs --follow
# clean up
copilot app delete
```
