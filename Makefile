.PHONY: gen
build:
	goctl api go --api ./dsl/auth.api --dir ./ --style goZero
	goctl model mysql ddl --src ./model/mysql/user.sql --dir ./model/mysql

.PHONY: build
build:
	go build -ldflags="-s -w" -o ./bin/auth-api ./api/auth

.PHONY: run
run:
	cd ./api && go run auth.go
