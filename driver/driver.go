package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	pb "github.com/wallaceicy06/muni-sign/proto"
)

const configFile = "/Users/sean/muni_sign_config.pb.txt"

var displayAddr = flag.String("display_addr", "raspberrypi.local:50051", "The display server address in the format of host:port")
var nextbusAddr = flag.String("nextbus_addr", "localhost:8081", "The nextbus server address in the format of host:port")
var adminAddr = flag.String("admin_addr", "http://localhost:8080", "The admin server address to use in the format http://host:port")

var colors = []*pb.Color{
	{
		Red:   1.0,
		Green: 0.0,
		Blue:  0.0,
	},
	{
		Red:   1.0,
		Green: 1.0,
		Blue:  0.0,
	},
	{
		Red:   0.0,
		Green: 1.0,
		Blue:  0.0,
	},
	{
		Red:   0.0,
		Green: 0.0,
		Blue:  1.0,
	},
	{
		Red:   0.6,
		Green: 0.0,
		Blue:  0.8,
	},
}

func main() {
	flag.Parse()

	dspConn, err := grpc.Dial(*displayAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting to display server: %v", err)
	}
	dspClient := pb.NewDisplayDriverClient(dspConn)

	nbConn, err := grpc.Dial(*nextbusAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error connecting to nextbus server: %v", err)
	}
	nbClient := pb.NewNextbusClient(nbConn)

	for {
		config, err := readConfigFile()
		if err != nil {
			log.Fatalf("Error reading configuration file: %v", err)
		}

		for i, stopId := range config.GetStopIds() {
			res, err := nbClient.ListPredictions(context.Background(), &pb.ListPredictionsRequest{
				Agency: config.GetAgency(),
				StopId: stopId,
			})
			if err != nil {
				log.Fatalf("Error listing predictions: %v", err)
			}

			for _, pred := range res.GetPredictions() {
				var msg string
				if l := len(pred.GetNextArrivals()); l == 1 {
					msg = fmt.Sprintf("%s-%s\n%d mins", pred.GetRoute(), pred.GetDestination(), pred.GetNextArrivals()[0])
				} else if l >= 2 {
					msg = fmt.Sprintf("%s-%s\n%d & %d mins", pred.GetRoute(), pred.GetDestination(), pred.GetNextArrivals()[0], pred.GetNextArrivals()[1])
				} else {
					continue
				}

				req := &pb.WriteRequest{
					Message: msg,
					Color:   colors[i%len(colors)],
				}

				if _, err := dspClient.Write(context.Background(), req); err != nil {
					log.Fatalf("Error writing: %v", err)
				}

				time.Sleep(time.Second * 5)
			}
		}
	}
}

func readConfigFile() (*pb.Configuration, error) {
	res, err := http.Get(fmt.Sprintf("%s/api/config", *adminAddr))
	if err != nil {
		return nil, fmt.Errorf("error getting config from admin server: %v", err)
	}

	parsedConfig := &pb.Configuration{}
	if err := json.NewDecoder(res.Body).Decode(parsedConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config proto: %v", err)
	}
	return parsedConfig, nil
}
