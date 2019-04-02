BUILD_VERSION := 0.0.1
BUILDOUT ?= chatbot
IMAGE_NAME = "iov1/chatbot:v${BUILD_VERSION}"

dist: deps clean test build

clean: 
	rm -rf $(BUILDOUT)

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) $(DOCKER_BUILD_FLAGS) -o $(BUILDOUT) .

test:
	go test -race ./...

deps:
	dep ensure --vendor-only

docker_build: deps
	docker build . -t $(IMAGE_NAME)

docker_push:
	docker push $(IMAGE_NAME)

docker: docker_build docker_push
