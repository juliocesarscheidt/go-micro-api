# azure variables
LOCATION?=eastus
RESOURCE_GROUP?=go-micro-api-rg
# backend variables
STORAGE_ACCOUNT_BACKEND?=gomicroapistorageacc
STORAGE_ACCOUNT_BACKEND_LOCATION?=eastus
# registry variables
REGISTRY_USERNAME?=gomicroapi
# api variables
API_NAME?=go-micro-api
API_VERSION?=v1.0.0
API_MESSAGE?=Hello World From ACI with Terraform

all: push-image init plan apply

create-resource-group:
	-@az group create --name $(RESOURCE_GROUP) --location $(LOCATION) 2> /dev/null

create-repo:
	-@az acr create --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --sku Basic
	-@az acr update --admin-enabled true --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME)

docker-login:
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) \
		--name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	echo "$${REGISTRY_PASSWORD}" | docker login "$(REGISTRY_USERNAME).azurecr.io" \
			--username "$(REGISTRY_USERNAME)" --password-stdin

push-image: create-resource-group create-repo docker-login
	docker image pull "docker.io/juliocesarmidia/$(API_NAME):v1.0.0"
	docker image tag "docker.io/juliocesarmidia/$(API_NAME):v1.0.0" "$(REGISTRY_USERNAME).azurecr.io/$(API_NAME):$(API_VERSION)"
	docker image push "$(REGISTRY_USERNAME).azurecr.io/$(API_NAME):$(API_VERSION)"

create-backend-storage-account:
	-@az storage account show --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) || \
		az storage account create --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) \
		--location $(STORAGE_ACCOUNT_BACKEND_LOCATION) --sku Standard_LRS
	-@az storage container create --resource-group $(RESOURCE_GROUP) \
		--account-name $(STORAGE_ACCOUNT_BACKEND) --name $(API_NAME) --public-access container

init: create-backend-storage-account
	terraform init -upgrade=true \
		-backend-config="resource_group_name=$(RESOURCE_GROUP)" \
		-backend-config="storage_account_name=$(STORAGE_ACCOUNT_BACKEND)" \
		-backend-config="container_name=$(API_NAME)" \
		-backend-config="key=terraform.tfstate"

tf-fmt:
	terraform fmt -write=true -recursive

tf-validate:
	terraform validate

tf-lint:
	docker container run --rm -t \
		--name tflint \
		--env TFLINT_LOG=debug \
		-v $$(pwd):/data \
		ghcr.io/terraform-linters/tflint

output:
	-@terraform output -raw app_gw_public_ip

plan: tf-validate tf-lint tf-fmt
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	terraform plan \
		-out=tfplan \
		-var-file=variables.tfvars \
		-var resource_group="$(RESOURCE_GROUP)" \
		-var registry_username="$(REGISTRY_USERNAME)" \
		-var registry_password="$${REGISTRY_PASSWORD}" \
		-var api_name="$(API_NAME)" \
		-var api_version="$(API_VERSION)" \
		-var api_message="$(API_MESSAGE)" \
		-input=false

apply:
	terraform apply tfplan

container-logs:
	az container logs --resource-group $(RESOURCE_GROUP) --name $(API_NAME) --follow

destroy:
	export REGISTRY_PASSWORD="$$(az acr credential show --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --query passwords[0].value | sed 's/^"//; s/"$$//')"; \
	terraform destroy \
		-var-file=variables.tfvars \
		-var resource_group="$(RESOURCE_GROUP)" \
		-var registry_username="$(REGISTRY_USERNAME)" \
		-var registry_password="$${REGISTRY_PASSWORD}" \
		-var api_name="$(API_NAME)" \
		-var api_version="$(API_VERSION)" \
		-var api_message="$(API_MESSAGE)" \
		-auto-approve
	make destroy-resources

destroy-resources:
	-@az acr delete --resource-group $(RESOURCE_GROUP) --name $(REGISTRY_USERNAME) --yes
	-@az storage account delete --resource-group $(RESOURCE_GROUP) --name $(STORAGE_ACCOUNT_BACKEND) --yes
	-@az group delete --name $(RESOURCE_GROUP) --yes
