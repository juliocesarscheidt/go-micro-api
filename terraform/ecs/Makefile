# AWS variables
AWS_ACCESS_KEY_ID?=
AWS_SECRET_ACCESS_KEY?=
AWS_DEFAULT_REGION?=us-east-1
ACCOUNT_ID?=
# backend variables
AWS_BACKEND_BUCKET?=gomicroapibucket
AWS_BACKEND_ACCESS_KEY_ID?=
AWS_BACKEND_SECRET_ACCESS_KEY?=
AWS_BACKEND_REGION?=us-east-1
# registry variables
REGISTRY_USERNAME?=AWS
REGISTRY_URL?=$(ACCOUNT_ID).dkr.ecr.$(AWS_DEFAULT_REGION).amazonaws.com
# api variables
API_NAME?=go-micro-api
API_VERSION?=v1.0.0
API_MESSAGE?=Hello World
API_ENVIRONMENT?=production

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## create-registry: Create ECR
.PHONY: create-registry
create-registry:
	-@aws ecr describe-repositories --repository-names $(API_NAME) --region $(AWS_DEFAULT_REGION) || \
		aws ecr create-repository --repository-name $(API_NAME) --region $(AWS_DEFAULT_REGION)

## docker-login: Login on ECR
.PHONY: docker-login
docker-login:
	aws ecr get-login-password --region $(AWS_DEFAULT_REGION) | \
		docker login --username $(REGISTRY_USERNAME) $(REGISTRY_URL) --password-stdin

## push-image: Push docker image to ECR
.PHONY: push-image
push-image: create-registry docker-login
	docker image pull "docker.io/juliocesarmidia/$(API_NAME):v1.0.0"
	docker image tag "docker.io/juliocesarmidia/$(API_NAME):v1.0.0" "$(REGISTRY_URL)/$(API_NAME):$(API_VERSION)"
	docker image push "$(REGISTRY_URL)/$(API_NAME):$(API_VERSION)"

## create-backend: Create the backend for Terraform
.PHONY: create-backend
create-backend:
	-@aws s3 ls s3://$(AWS_BACKEND_BUCKET) --region $(AWS_BACKEND_REGION) 2> /dev/null || \
		aws s3api create-bucket --bucket $(AWS_BACKEND_BUCKET) --region $(AWS_BACKEND_REGION) --acl private

## tf-init: Initialization of Terraform
.PHONY: tf-init
tf-init: create-backend
	terraform init -upgrade=true \
		-backend-config="bucket=$(AWS_BACKEND_BUCKET)" \
		-backend-config="key=state.tfstate" \
		-backend-config="region=$(AWS_BACKEND_REGION)" \
		-backend-config="workspace_key_prefix=terraform/$(API_NAME)" \
		-backend-config="access_key=$(AWS_BACKEND_ACCESS_KEY_ID)" \
		-backend-config="secret_key=$(AWS_BACKEND_SECRET_ACCESS_KEY)" \
		-backend-config="encrypt=true"

## tf-fmt: Format Terraform scripts
.PHONY: tf-fmt
tf-fmt:
	terraform fmt -write=true -recursive

## tf-validate: Validate Terraform scripts
.PHONY: tf-validate
tf-validate:
	terraform validate

## tf-lint: Lint Terraform scripts
.PHONY: tf-lint
tf-lint:
	docker container run --rm -t \
		--name tflint \
		--env TFLINT_LOG=debug \
		-v $$(pwd):/data \
		ghcr.io/terraform-linters/tflint

## tf-plan: Planning of Terraform
.PHONY: tf-plan
tf-plan: tf-validate tf-lint tf-fmt
	terraform plan \
		-out=tfplan \
		-var-file=variables.tfvars \
		-var aws_region="$(AWS_DEFAULT_REGION)" \
		-var registry_url="$(REGISTRY_URL)" \
		-var api_name="$(API_NAME)" \
		-var api_version="$(API_VERSION)" \
		-var api_message="$(API_MESSAGE)" \
		-var api_environment="$(API_ENVIRONMENT)" \
		-input=false

## tf-apply: Apply of Terraform
.PHONY: tf-apply
tf-apply:
	terraform apply tfplan

## tf-output: Output of Terraform
.PHONY: tf-output
tf-output:
	-@terraform output -raw public_ip

## tf-refresh: Refresh Terraform state
.PHONY: tf-refresh
tf-refresh:
	terraform refresh -var-file=variables.tfvars

## logs: Logs from ECS
.PHONY: logs
logs:
	export QUERY_ID=$$(aws logs start-query --region $(AWS_DEFAULT_REGION) \
		--log-group-name "/aws/ecs/$(API_NAME)" \
		--start-time "$$(date -d "$$(date '+%Y-%m-%d') 00:00:00" +%s)" \
		--end-time "$$(date -d "$$(date '+%Y-%m-%d') 23:59:59" +%s)" \
		--query queryId --output text \
		--query-string 'fields host, ip, message as msg, method, path, severity, timestamp | sort @timestamp desc | limit 100'); \
	sleep 5; \
	aws logs get-query-results --region $(AWS_DEFAULT_REGION) \
		--no-paginate --query results \
		--output text --query-id "$$QUERY_ID"

## tf-destroy: Destroy Terraform resources and then aws resources
.PHONY: tf-destroy
tf-destroy:
	terraform destroy \
		-var-file=variables.tfvars \
		-var aws_region="$(AWS_DEFAULT_REGION)" \
		-var registry_url="$(REGISTRY_URL)" \
		-var api_name="$(API_NAME)" \
		-var api_version="$(API_VERSION)" \
		-var api_message="$(API_MESSAGE)" \
		-var api_environment="$(API_ENVIRONMENT)" \
		-auto-approve
	make destroy-base-resources

## destroy-base-resources: Destroy base resources
.PHONY: destroy-base-resources
destroy-base-resources:
	-@aws s3 rm s3://$(AWS_BACKEND_BUCKET)/ --recursive
	-@aws s3api delete-bucket --bucket $(AWS_BACKEND_BUCKET) --region $(AWS_BACKEND_REGION)
	-@aws ecr batch-delete-image --repository-name $(API_NAME) --region $(AWS_DEFAULT_REGION) --image-ids imageTag=v1.0.0
	-@aws ecr delete-repository --repository-name $(API_NAME) --region $(AWS_DEFAULT_REGION)
