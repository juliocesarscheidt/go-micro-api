# azure variables
LOCATION?=eastus
RESOURCE_GROUP?=go-micro-api-rg
# backend variables
STORAGE_ACCOUNT_BACKEND?=gomicroapistorageacc
STORAGE_ACCOUNT_BACKEND_LOCATION?=eastus
# registry variables
REGISTRY_USERNAME?=gomicroapi
REGISTRY_URL?=$(REGISTRY_USERNAME).azurecr.io
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

## create-resource-group: Create resource group
.PHONY: create-resource-group
create-resource-group:
	-@az group create --name $(RESOURCE_GROUP) --location $(LOCATION) 2> /dev/null

## create-registry: Create ACR
.PHONY: create-registry
create-registry:
	-@az acr create --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --sku Basic
	-@az acr update --admin-enabled true --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME)

## docker-login: Login on ACR
.PHONY: docker-login
docker-login:
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) \
		--name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	echo "$${REGISTRY_PASSWORD}" | docker login "$(REGISTRY_URL)" \
			--username "$(REGISTRY_USERNAME)" --password-stdin

## push-image: Push docker image to ACR
.PHONY: push-image
push-image: create-resource-group create-registry docker-login
	docker image pull "docker.io/juliocesarmidia/$(API_NAME):v1.0.0"
	docker image tag "docker.io/juliocesarmidia/$(API_NAME):v1.0.0" "$(REGISTRY_URL)/$(API_NAME):$(API_VERSION)"
	docker image push "$(REGISTRY_URL)/$(API_NAME):$(API_VERSION)"

## create-backend: Create the backend for Terraform
.PHONY: create-backend
create-backend:
	-@az storage account show --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) 2> /dev/null || \
		az storage account create --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) \
		--location $(STORAGE_ACCOUNT_BACKEND_LOCATION) --sku Standard_LRS
	-@az storage container create --resource-group $(RESOURCE_GROUP) \
		--account-name $(STORAGE_ACCOUNT_BACKEND) --name $(API_NAME) --public-access container

## tf-init: Initialization of Terraform
.PHONY: tf-init
tf-init: create-backend
	terraform init -upgrade=true \
		-backend-config="resource_group_name=$(RESOURCE_GROUP)" \
		-backend-config="storage_account_name=$(STORAGE_ACCOUNT_BACKEND)" \
		-backend-config="container_name=$(API_NAME)" \
		-backend-config="key=state.tfstate"

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
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) \
		--name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	terraform plan \
		-out=tfplan \
		-var-file=variables.tfvars \
		-var resource_group="$(RESOURCE_GROUP)" \
		-var registry_username="$(REGISTRY_USERNAME)" \
		-var registry_password="$${REGISTRY_PASSWORD}" \
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

## logs: Logs from container instance
.PHONY: logs
logs:
	az container logs --resource-group $(RESOURCE_GROUP) --name $(API_NAME) --follow

## tf-destroy: Destroy Terraform resources and then Azure resources
.PHONY: tf-destroy
tf-destroy:
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) \
		--name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	terraform destroy \
		-var-file=variables.tfvars \
		-var resource_group="$(RESOURCE_GROUP)" \
		-var registry_username="$(REGISTRY_USERNAME)" \
		-var registry_password="$${REGISTRY_PASSWORD}" \
		-var api_name="$(API_NAME)" \
		-var api_version="$(API_VERSION)" \
		-var api_message="$(API_MESSAGE)" \
		-var api_environment="$(API_ENVIRONMENT)" \
		-auto-approve
	make destroy-base-resources

## destroy-base-resources: Destroy base resources
.PHONY: destroy-base-resources
destroy-base-resources:
	-@az acr delete --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --yes
	-@az storage account delete --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) --yes
	-@az group delete --name $(RESOURCE_GROUP) --yes
