#!/bin/bash

# Auth Service 测试脚本

echo "================================"
echo "Auth Service 单元测试"
echo "================================"

cd "$(dirname "$0")/.."

# 统计变量
TOTAL=0
PASS=0
FAIL=0
FAILED_LIST=""

run_test() {
    local NAME=$1
    local TARGET=$2
    ((TOTAL++))
    
    echo ""
    echo "Running: $NAME..."
    
    # 运行测试，捕获输出和退出码
    OUTPUT=$(go test -v "$TARGET" 2>&1)
    EXIT_CODE=$?
    
    # 输出过滤后的日志
    echo "$OUTPUT" | grep -E "(PASS|FAIL|RUN|ok|===)"
    
    if [ $EXIT_CODE -eq 0 ]; then
        ((PASS++))
        echo "✅ PASS"
    else
        ((FAIL++))
        echo "❌ FAIL"
        FAILED_LIST+="$FAILED_LIST\n- $NAME ($TARGET)"
    fi
}

# 1. 运行错误定义测试
run_test "错误定义测试" "./tests/unit/types/errors_test.go"

# 2. 运行类型测试
run_test "类型测试" "./tests/unit/types/types_test.go"

# 3. 运行密码加密测试
run_test "密码加密测试" "./tests/unit/svc/password_test.go"

# 4. 运行JWT测试
run_test "JWT测试" "./tests/unit/svc/jwt_test.go"

# 5. 运行登录逻辑测试
run_test "登录逻辑测试" "./tests/unit/logic/logic_login_test.go"

# 6. 运行注册逻辑测试
run_test "注册逻辑测试" "./tests/unit/logic/logic_register_test.go"

echo ""
echo "================================"
echo "测试结果汇总"
echo "================================"
echo "总用例数: $TOTAL"
echo "通过    : $PASS"
echo "失败    : $FAIL"

if [ $TOTAL -gt 0 ]; then
    RATE=$(awk "BEGIN {printf \"%.2f\", ($PASS/$TOTAL)*100}")
    echo "通过率  : ${RATE}%"
fi

if [ $FAIL -gt 0 ]; then
    echo ""
    echo -e "以下测试未通过:$FAILED_LIST"
    exit 1
else
    echo ""
    echo "✅ 所有测试通过!"
    exit 0
fi
