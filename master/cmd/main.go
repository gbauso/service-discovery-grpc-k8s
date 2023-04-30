package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	"github.com/gbauso/service-discovery-grpc-k8s/master/repository"
	abstraction "github.com/gbauso/service-discovery-grpc-k8s/master/repository/interface"
	"github.com/gbauso/service-discovery-grpc-k8s/master/server"
	"github.com/gbauso/service-discovery-grpc-k8s/master/server/interceptors"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/nu7hatch/gouuid"
)

var (
	port         = flag.Int("port", 50058, "The server port")
	logPath      = flag.String("log-path", "/tmp/discovery_master-%.log", "Log path")
	databasePath = flag.String("db-path", "../database/sqlite.db", "SQLite file path")
)

func main() {
	log := logger.New()
	id, _ := uuid.NewV4()
	log.SetFormatter(&logger.JSONFormatter{})
	flag.Parse()
	fileName := fmt.Sprintf(*logPath, id)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("error opening file: %v", err)
		panic(err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)

	db, err := sql.Open("sqlite3", *databasePath)
	if err != nil {
		log.Errorf("error opening database: %v", err)
		panic(err)
	}

	var repo abstraction.ServiceHandlerRepository := repository.NewServiceHandlerRepository(db, context.Background())

	server := server.NewServer(repo)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Errorf("failed to listen")
		panic(err)
	}

	loggingInterceptor := interceptors.NewLoggingInterceptor(log)

	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor.ServerInterceptor))
	pb.RegisterDiscoveryServiceServer(s, server)
	reflection.Register(s)

	log.Infof("server listenning on 0.0.0.0:%d", port)
	s.Serve(lis)
}
