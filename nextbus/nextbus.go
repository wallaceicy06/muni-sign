package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	nb "github.com/dinedal/nextbus"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type nextbus interface {
	GetAgencyList() ([]nb.Agency, error)
	GetStopPredictions(agencyTag string, stopID string) ([]nb.PredictionData, error)
}

type server struct {
	nbClient nextbus
	port     int
}

var port = flag.Int("port", 8081, "the port to host the nextbus server on")

func main() {
	flag.Parse()

	srv := newServer(*port, nb.DefaultClient).serve()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	srv.GracefulStop()
	os.Exit(0)
}

func newServer(port int, nbClient nextbus) *server {
	return &server{
		port:     port,
		nbClient: nbClient,
	}
}

func (s *server) serve() *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterNextbusServer(grpcSrv, s)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc server stopped: %v", err)
		}
	}()

	return grpcSrv
}

func (s *server) ListAgencies(ctx context.Context, req *pb.ListAgenciesRequest) (*pb.ListAgenciesResponse, error) {
	agencies, err := s.nbClient.GetAgencyList()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Problem getting agency list: %v", err)
	}

	res := &pb.ListAgenciesResponse{}
	for _, a := range agencies {
		res.Agencies = append(res.Agencies, &pb.Agency{Tag: a.Tag, Name: a.Title})
	}

	return res, nil
}

func (s *server) ListPredictions(ctx context.Context, req *pb.ListPredictionsRequest) (*pb.ListPredictionsResponse, error) {
	if req.Agency == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "Agency is required.")
	}
	if req.StopId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "StopID is required.")
	}

	preds, err := s.nbClient.GetStopPredictions(req.Agency, req.StopId)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Problem getting predictions: %v", err)
	}

	res := &pb.ListPredictionsResponse{}

	for _, pred := range preds {
		for _, dir := range pred.PredictionDirectionList {
			if len(dir.PredictionList) == 0 {
				continue
			}

			p := &pb.Prediction{Route: pred.RouteTag, Destination: dir.Title}
			for _, n := range dir.PredictionList {
				mins, err := strconv.Atoi(n.Minutes)
				if err != nil {
					return nil, grpc.Errorf(codes.Internal, "Problem converting string to integer: %v", err)
				}
				p.NextArrivals = append(p.NextArrivals, int32(mins))
			}

			res.Predictions = append(res.Predictions, p)
		}
	}

	return res, nil
}
