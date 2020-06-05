REGISTRIES?=docker.io
REPOSITORY?=yzihua
APP=hpa-exporter
V=$(shell cat VERSION)


build:
	GOOS=linux GOARCH=amd64 go build -o deploy/bin/hpa-exporter main.go


image: build
	@docker build -f deploy/Dockerfile deploy -t $(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)
	@docker push $(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)
	@echo "$(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)"
