package main

import (
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"

	"google.golang.org/grpc"

	"github.com/spf13/viper"
)

var grpcConn *grpc.ClientConn
var accountServiceClient sso.AccountServiceClient
var authorizationServiceClient sso.AuthorizationServiceClient

func startGrpcClient() {
	var err error
	grpcConn, err = grpc.Dial(viper.GetString("grpcServerAddr"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Fail to dial gRPC server: %v", err)
	}
	accountServiceClient = sso.NewAccountServiceClient(grpcConn)
	authorizationServiceClient = sso.NewAuthorizationServiceClient(grpcConn)
}

func stopGrpcClient() {
	log.Info("Stopping gRPC client")
	err := grpcConn.Close()
	if err != nil {
		log.Errorf("Error while shutting down gRPC client: %v", err)
	}
}
