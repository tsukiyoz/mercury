.PHONY: docker
docker:
	@rm mercury || true
	@docker rmi -f mercury:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o mercury .
	@docker build -t mercury:v0.0.1 .

docker-k8s:
	@rm mercury || true
	@docker rmi -f mercury:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags k8s -o mercury .
	@docker build -t mercury:v0.0.1 .

.PHONY: k8s-setup-db
k8s-setup-db:
	@kubectl apply -f k8s-mysql.yaml
	@kubectl apply -f k8s-redis.yaml

.PHONY: k8s-mysql-init
k8s-mysql-init:
	@cat script/mysql/init.sql | mysql -h 127.0.0.1 -P 3308 -u root -p'for.nothing'

.PHONY: k8s-setup-web
k8s-setup-web:
	@kubectl apply -f k8s-mercury.yaml
	@kubectl apply -f k8s-ingress-nginx.yaml

.PHONY: k8s-teardown-web
k8s-teardown-web:
	@kubectl delete ing mercury-ingress || true
	@kubectl delete deployment mercury-deployment || true

.PHONY: k8s-teardown-db
k8s-teardown-db:
	@kubectl delete deployment mysql-deployment || true
	@kubectl delete deployment redis-deployment || true

.PHONY: k8s-teardown
k8s-teardown:
	make k8s-teardown-web
	make k8s-teardown-db

.PHONY: k8s-reload-web
k8s-reload-web:
	make k8s-teardown-web
	make mock
	make docker-k8s
	make k8s-setup-web

.PHONY: mock
mock:
	@go generate ./...
	@go mod tidy

.PHONY: grpc
grpc:
	@buf generate api/proto
