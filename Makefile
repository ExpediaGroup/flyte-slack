test:
	go build && go test ./...
build: test
	go build .
docker-build:
	docker build -t flyte-slack .
