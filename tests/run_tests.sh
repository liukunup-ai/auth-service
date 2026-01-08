#!/bin/bash

# Auth Service 测试脚本

echo "================================"
echo "Auth Service 单元测试"
echo "================================"

cd "$(dirname "$0")/.."

echo ""
echo "1. 运行错误定义测试..."
go test -v ./tests/errors_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "2. 运行类型测试..."
go test -v ./tests/types_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "3. 运行密码加密测试..."
go test -v ./tests/password_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "4. 运行JWT测试..."
go test -v ./tests/jwt_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "5. 运行登录逻辑测试..."
go test -v ./tests/logic_login_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "6. 运行注册逻辑测试..."
go test -v ./tests/logic_register_test.go 2>&1 | grep -E "(PASS|FAIL|RUN|ok|===)" || true

echo ""
echo "================================"
echo "测试完成"
echo "================================"
