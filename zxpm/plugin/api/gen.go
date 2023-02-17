package api

//go:generate protoc -I ./ ./command.proto ./config.proto ./task-interface.proto --go_out=.. --go-grpc_out=..
