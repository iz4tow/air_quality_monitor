#!/bin/bash
sudo apt-get install gcc-arm* -y
export GOOS=linux; \
export GOARCH=arm; \
export GOARM=7; \
export CC=arm-linux-gnueabi-gcc; \
CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" --trimpath \
 data_logger.go
