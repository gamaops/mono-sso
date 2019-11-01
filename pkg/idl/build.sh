#!/bin/bash

protoc -I ./ --go_out=paths=source_relative,plugins=grpc:. ./sso-common/*.proto
protoc -I ./ --go_out=paths=source_relative,plugins=grpc:.  ./sso-service/*.proto
protoc -I ./ --go_out=paths=source_relative,plugins=grpc:. ./sso-manager/*.proto
cd sso-common
go install
cd ../sso-manager
go install
cd ../sso-service
go install
cd ../