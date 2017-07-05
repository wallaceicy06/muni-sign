package main

import (
	"fmt"
	"testing"

	"google.golang.org/grpc"

	nb "github.com/dinedal/nextbus"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type fakeNextbus struct {
}

func (fnb *fakeNextbus) GetStopPredictions(agencyTag string, stopID string) ([]nb.PredictionData, error) {
	return nil, nil
}

const testPort = 25565

func TestServing(t *testing.T) {
	srv := newServer(testPort, &fakeNextbus{}).serve()
	defer srv.Stop()

	conn, err := grpc.Dial(fmt.Sprintf(":%d", testPort), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Could not reach Nextbus RPC server.")
	}
	defer conn.Close()
}
