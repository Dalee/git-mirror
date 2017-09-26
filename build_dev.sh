#!/usr/bin/env bash

GOOS=linux GOARCH=amd64 go build git-mirror.go config.go
