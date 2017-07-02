package config

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"

	pb "github.com/wallaceicy06/sign-server/proto"
)

const configFile = "/Users/sean/muni_sign_config.pb.txt"

func GetConfig() (*pb.Configuration, error) {
	config, err := readConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error getting configuration: %v", err)
	}
	return config, nil
}

func SetConfig(newConfig *pb.Configuration) {
}

func readConfigFile() (*pb.Configuration, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	parsedConfig := &pb.Configuration{}
	if err := proto.UnmarshalText(string(data), parsedConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config proto: %v", err)
	}
	return parsedConfig, nil
}
