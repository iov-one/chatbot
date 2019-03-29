VERSION := 0.0.1

docker_build:
	docker build . -t iov1/chatbot:$(VERSION)

docker_push:
	docker push iov1/chatbot:$(VERSION)

docker: docker_build docker_push