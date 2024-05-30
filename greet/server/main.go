package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/smohandoss0611/gRPC-golang/greet/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var greetWithDeadlineTime time.Duration = 1 * time.Second

var addr string = "0.0.0.0:50051"

type Server struct {
	pb.GreetServiceServer
}

func CheckHeaderInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		headers, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return nil, status.Error(codes.Internal, "Error while reading the context")
		}

		if len(headers.Get("authorization")) == 0 {
			return nil, status.Error(codes.Unauthenticated, "Expected authorization header")
		}

		return handler(ctx, req)
	}
}

func LogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("Received a request: %v\n", req)

		headers, ok := metadata.FromIncomingContext(ctx)

		if ok {
			log.Printf("Received headers: %v\n", headers)
		}

		return handler(ctx, req)
	}
}

func main() {
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	defer lis.Close()
	log.Printf("Listening at %s\n", addr)

	opts := []grpc.ServerOption{}

	tls := false // change that to true if needed
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)

		if err != nil {
			log.Fatalf("Failed loading certificates: %v\n", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	opts = append(opts, grpc.ChainUnaryInterceptor(LogInterceptor(), CheckHeaderInterceptor()))

	s := grpc.NewServer(opts...)
	pb.RegisterGreetServiceServer(s, &Server{})

	defer s.Stop()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
