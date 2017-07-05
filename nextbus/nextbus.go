package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	nb "github.com/dinedal/nextbus"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type server struct{}

var port = flag.Int("port", 8081, "the port to host the nextbus server on")

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNextbusServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) ListPredictions(ctx context.Context, req *pb.ListPredictionsRequest) (*pb.ListPredictionsResponse, error) {
	if req.Agency == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "Agency is required.")
	}
	if req.StopId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "StopID is required.")
	}

	preds, err := nb.DefaultClient.GetStopPredictions(req.Agency, req.StopId)
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
