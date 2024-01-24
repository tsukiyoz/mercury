.PHONY: docker
docker:
	@rm webook || true
	@docker rmi -f webook:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker build -t webook:v0.0.1 .

docker-k8s:
	@rm webook || true
	@docker rmi -f webook:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags k8s -o webook .
	@docker build -t webook:v0.0.1 .

.PHONY: k8s-setup-db
k8s-setup-db:
	@kubectl apply -f k8s-mysql.yaml
	@kubectl apply -f k8s-redis.yaml

.PHONY: k8s-mysql-init
k8s-mysql-init:
	@cat script/mysql/init.sql | mysql -h 127.0.0.1 -P 3308 -u root -p'for.nothing'

.PHONY: k8s-setup-web
k8s-setup-web:
	@kubectl apply -f k8s-webook.yaml
	@kubectl apply -f k8s-ingress-nginx.yaml

.PHONY: k8s-teardown-web
k8s-teardown-web:
	@kubectl delete ing webook-ingress || true
	@kubectl delete deployment webook-deployment || true

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
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user.mock.go
	@mockgen -source=internal/service/article.go -package=svcmocks -destination=internal/service/mocks/article.mock.go
	@mockgen -source=internal/service/captcha.go -package=svcmocks -destination=internal/service/mocks/captcha.mock.go

	@mockgen -source=internal/repository/user.go -package=repomocks -destination=internal/repository/mocks/user.mock.go
	@mockgen -source=internal/repository/captcha.go -package=repomocks -destination=internal/repository/mocks/captcha.mock.go
	@mockgen -source=internal/repository/article/article.go -package=articlerepomocks -destination=internal/repository/mocks/article/article.mock.go
	@mockgen -source=internal/repository/article/article_author.go -package=articlerepomocks -destination=internal/repository/mocks/article/article_author.mock.go
	@mockgen -source=internal/repository/article/article_reader.go -package=articlerepomocks -destination=internal/repository/mocks/article/article_reader.mock.go

	@mockgen -source=internal/repository/dao/user.go -package=daomocks -destination=internal/repository/mocks/dao/user.mock.go
	@mockgen -source=internal/repository/cache/user/user.go -package=cachemocks -destination=internal/repository/mocks/cache/user/user.mock.go

	@mockgen -source=pkg/ratelimit/types.go -package=limitmocks -destination=pkg/ratelimit/mocks/ratelimit.mock.go

	@mockgen -package=redismocks -destination=internal/repository/mocks/cache/redis/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable

	@go mod tidy
