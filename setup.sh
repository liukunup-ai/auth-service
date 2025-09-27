#!/bin/bash

set -euo pipefail

sudo sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g'      /etc/apt/sources.list.d/debian.sources
sudo sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list.d/debian.sources

sudo apt-get update
sudo apt-get install -y \
    curl \
    wget \
    vim

sudo rm -rf /var/lib/apt/lists/*

go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=on
go env -w GONOPROXY=none

go install github.com/go-delve/delve/cmd/dlv@latest

go install github.com/zeromicro/go-zero/tools/goctl@latest

# goctl env check --install --verbose --force
# go mod init auth-service
# go get -u github.com/zeromicro/go-zero@latest
