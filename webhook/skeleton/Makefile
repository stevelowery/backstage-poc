NAMESPACE := ${{ values.namespace }}
NAME      := ${{ values.component_id }}
VERSION   := v1.21.8

##########################
# Local tasks
##########################
bootstrap: init-local init install install-webhook install-examples

init-local:
	@helm repo up
	@minikube start -p ${VERSION} --kubernetes-version=${VERSION}
	@minikube profile ${VERSION}
	@kubectl create ns ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
	@kubectl create ns nginx-ingress --dry-run=client -o yaml | kubectl apply -f -
	@kubectl create deployment nginx-ingress-internal-controller --image httpd --replicas 1 --namespace nginx-ingress --dry-run=client -o yaml | kubectl apply -f -

init:
	@kubectl create secret tls ${NAME}-tls -n ${NAMESPACE} --cert=certs/tls.crt --key=certs/tls.key --dry-run=client -o yaml | kubectl apply -n ${NAMESPACE} -f -
	@kubectl apply -f conf/manifests/role-and-binding.yaml

generate:
	@go generate ./...

build:
	@go build -o bin/app main.go

test:
	@go test ./... -cover

docker: build
	@docker build -t us-central1-docker.pkg.dev/us-mgmt/entegral-official/${NAMESPACE}/${NAME}:latest .
	@docker push us-central1-docker.pkg.dev/us-mgmt/entegral-official/${NAMESPACE}/${NAME}:latest

docker-mk: build
	@eval $$(minikube docker-env) && docker build -t us-central1-docker.pkg.dev/us-mgmt/entegral-official/${NAMESPACE}/${NAME}:latest .

install: docker
	@helm upgrade ${NAME} abs/smart-chart \
	--install --atomic --namespace=${NAMESPACE} --create-namespace \
	--values=conf/helm/values.yaml

install-mk: docker-mk
	@helm upgrade ${NAME} abs/smart-chart \
	--install --atomic --namespace=${NAMESPACE} --create-namespace \
	--values=conf/helm/values.yaml --values=conf/helm/values.minikube.yaml

uninstall:
	@helm uninstall ${NAME} --namespace=${NAMESPACE}

install-webhook:
	@helm upgrade ${NAME}-config abs/webhook --install --namespace=${NAMESPACE} --atomic --values conf/helm/values.webhook.yaml \
	 --set clientConfig.caBundle="$$(cat certs/ca.crt)"

uninstall-webhook:
	@helm uninstall ${NAME}-config --namespace=${NAMESPACE}

install-examples:
	@for f in $$(find examples -type f); do \
		kubectl apply -f $$f --force; \
	done

uninstall-examples:
	@for f in $$(find examples -type f); do \
		kubectl delete -f $$f; \
	done

##########################
# CI tasks
##########################
build-ci:
	@jfrog rt go build -o bin/app .

test-ci:
	@go test -race -v -coverprofile=coverage.txt ./...
