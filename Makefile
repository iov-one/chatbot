BUILD_VERSION ?= manual
BUILDOUT ?= chatbot
IMAGE_NAME = "iov1ops/chatbot:${BUILD_VERSION}"

dist: deps clean test build

clean: 
	rm -rf $(BUILDOUT)

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) $(DOCKER_BUILD_FLAGS) -o $(BUILDOUT) .

test:
	go test -race ./...

deps:
ifndef $(shell command -v dep help > /dev/null)
	go get github.com/golang/dep/cmd/dep
endif
	dep ensure -vendor-only

docker_build: deps
	docker build . -t $(IMAGE_NAME)

docker_push:
	docker push $(IMAGE_NAME)

docker: docker_build docker_push
