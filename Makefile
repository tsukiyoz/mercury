.PHONY: docker
docker:
	@rm webook || true
	@docker rmi -f webook:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker build -t webook:v0.0.1 .