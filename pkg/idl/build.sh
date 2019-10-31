#!/bin/bash

protoc -I ./ --go_out=plugins=grpc:. ./sso-common/*.proto
protoc -I ./ --go_out=plugins=grpc:. ./sso-service/*.proto
protoc -I ./ --go_out=plugins=grpc:. ./sso-manager/*.proto