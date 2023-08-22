.PHONY: docker
docker:
	@rm webook || true
	@docker rmi -f webook:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker build -t webook:v0.0.1 .

.PHONY: k8s-setup-all
k8s-setup-all:
	@kubectl apply -f k8s-mysql.yaml
	@kubectl apply -f k8s-redis.yaml
	@kubectl apply -f k8s-webook.yaml
	@kubectl apply -f k8s-ingress-nginx.yaml

.PHONY: k8s-teardown
k8s-teardown:
	@kubectl delete ing webook-ingress
	@kubectl delete deployment webook-deployment
	@kubectl delete deployment mysql-deployment
	@kubectl delete deployment redis-deployment