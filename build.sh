#!/bin/env bash
docker run --rm -v "$PWD":/root -w /root -e GOOS=linux -e GOARCH=arm -e GOARM=5 golang:1.19.11-alpine3.18 go build -v
