.PHONY: init
init:
	go install github.com/go-delve/delve/cmd/dlv@latest
	go install github.com/zeromicro/go-zero/tools/goctl@latest
# 	goctl env check --install --verbose --force
# 	go mod init auth-service
# 	go get -u github.com/zeromicro/go-zero@latest

.PHONY: gen
build:
	goctl api go --api ./api/dsl/auth.api --dir ./api/ --style goZero
	goctl rpc protoc ./rpc/dsl/auth.proto --go_out=./rpc --go-grpc_out=./rpc --zrpc_out=./rpc --style goZero
	goctl model mysql ddl --src ./model/mysql/user.sql --dir ./model/mysql

.PHONY: build
build:
	go build -ldflags="-s -w" -o ./bin/auth-api ./api/auth
# 	go build -ldflags="-s -w" -o ./bin/auth-rpc ./rpc/auth

.PHONY: run
run:
	cd ./api && go run auth.go
