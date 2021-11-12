package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/davebehr1/microservices/volume-service/proto/volume"

	"github.com/davebehr1/microservices/volume-service/constants"

	"github.com/davebehr1/microservices/volume-service/volume"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type service struct {
	pb.UnimplementedVolumeServiceServer
}

func (s *service) Add(ctx context.Context, req *pb.AddParams) (*pb.Response, error) {

	return &pb.Response{Total: req.NumOne + req.NumTwo}, nil
}

func main() {
	log.Print("starting VOLUME microservice")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterVolumeServiceServer(s, &service{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	temporalClient := initTemporalClient()

	worker := initActivityWorker(temporalClient)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-signals

	worker.Stop()

	log.Print("closing VOLUME microservice")
}

func initTemporalClient() client.Client {
	temporalClientOptions := client.Options{HostPort: net.JoinHostPort("localhost", "7233")}
	temporalClient, err := client.NewClient(temporalClientOptions)
	if err != nil {
		log.Fatal("cannot start temporal client: " + err.Error())
	}
	return temporalClient
}

func initActivityWorker(temporalClient client.Client) worker.Worker {
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize: constants.MaxConcurrentVolumeActivitySize,
	}
	worker := worker.New(temporalClient, constants.VolumeActivityQueue, workerOptions)
	worker.RegisterActivity(volume.Service{}.CalculateParallelepipedVolume)

	err := worker.Start()
	if err != nil {
		log.Fatal("cannot start temporal worker: " + err.Error())
	}
	return worker
}
