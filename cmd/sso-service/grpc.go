package main

import (
	"net"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var gRPCServer *grpc.Server

func startGrpcServer() {

	log.Info("Starting gRPC Server")

	lis, err := net.Listen("tcp", viper.GetString("grpcListen"))

	if err != nil {
		log.Fatalf("Error while listening on gRPC server address: %v", err)
	}

	accountSrv := &AccountServer{}
	authorizationSrv := &AuthorizationServer{}
	sessionSrv := &SessionServer{}
	gRPCServer = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: viper.GetDuration("grpcMaxKeepAlive")}),
	)

	sso.RegisterAccountServiceServer(gRPCServer, accountSrv)
	sso.RegisterAuthorizationServiceServer(gRPCServer, authorizationSrv)
	sso.RegisterSessionServiceServer(gRPCServer, sessionSrv)
	gRPCServer.Serve(lis)

}

func stopGrpcServer() {
	log.Warn("Stopping gRPC server")
	gRPCServer.GracefulStop()
}
