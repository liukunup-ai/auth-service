.PHONY: init
init:
	go install github.com/go-delve/delve/cmd/dlv@latest
	go install github.com/zeromicro/go-zero/tools/goctl@latest
# 	goctl env check --install --verbose --force
# 	go mod init auth-service
# 	go get -u github.com/zeromicro/go-zero@latest

# .PHONY: gen
# build:
# 	goctl api go -api ./api/auth.api -dir ./api -style goZero

.PHONY: build
build:
	go build -ldflags="-s -w" -o ./bin/auth-api ./api/auth
# 	go build -ldflags="-s -w" -o ./bin/auth-rpc ./rpc/auth