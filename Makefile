BUILD_VERSION ?= manual
BUILDOUT ?= chatbot
IMAGE_NAME = "iov1ops/chatbot:${BUILD_VERSION}"
# make sure we turn on go modules
export GO111MODULE := on

dist: deps clean test build

clean: 
	rm -rf $(BUILDOUT)

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) $(DOCKER_BUILD_FLAGS) -o $(BUILDOUT) .

test:
	go test -race ./...
deps:
	@ go mod vendor

docker_build: deps
	docker build . -t $(IMAGE_NAME)

docker_push:
	docker push $(IMAGE_NAME)

docker: docker_build docker_push
