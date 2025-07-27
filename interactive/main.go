package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	intrv1 "red-feed/api/proto/gen/intr/v1"
	grpcintr "red-feed/interactive/grpc"
)

func main() {
	server := grpc.NewServer()
	intrSvc := &grpcintr.InteractiveServiceServer{}
	intrv1.RegisterInteractiveServiceServer(server, intrSvc)
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	// 这边会阻塞，类似与 gin.Run
	err = server.Serve(l)
	log.Println(err)
}
