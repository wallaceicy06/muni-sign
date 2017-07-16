package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	nb "github.com/dinedal/nextbus"
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type fakeNextbus struct {
	agencyList []nb.Agency
	agencyErr  error
}

func (fnb *fakeNextbus) GetAgencyList() ([]nb.Agency, error) {
	if fnb.agencyErr != nil {
		return nil, fnb.agencyErr
	}
	return fnb.agencyList, nil
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

func TestListAgencies(t *testing.T) {
	var testAgencies = []nb.Agency{
		{Tag: "sf-muni", Title: "San Francisco MTA"},
		{Tag: "la-metro", Title: "Los Angeles MTA"},
	}

	tests := []struct {
		name     string
		fakeNb   *fakeNextbus
		wantRes  *pb.ListAgenciesResponse
		wantCode codes.Code
	}{
		{
			name:   "Good",
			fakeNb: &fakeNextbus{agencyList: testAgencies},
			wantRes: &pb.ListAgenciesResponse{Agencies: []*pb.Agency{
				{Tag: "sf-muni", Name: "San Francisco MTA"},
				{Tag: "la-metro", Name: "Los Angeles MTA"},
			}},
			wantCode: codes.OK,
		},
		{
			name:     "Error",
			fakeNb:   &fakeNextbus{agencyErr: errors.New("fake agency list error")},
			wantCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			srv := newServer(testPort, test.fakeNb)

			gotRes, err := srv.ListAgencies(ctx, &pb.ListAgenciesRequest{})

			if gotCode := grpc.Code(err); gotCode != test.wantCode {
				t.Errorf("ListAgencies(_, _) got code %d want %d", gotCode, test.wantCode)
				return
			}

			if test.wantCode != codes.OK {
				return
			}

			if !proto.Equal(gotRes, test.wantRes) {
				t.Errorf("ListAgencies(_, _) = %v, _ want %v, _", gotRes, test.wantRes)
			}
		})
	}
}
