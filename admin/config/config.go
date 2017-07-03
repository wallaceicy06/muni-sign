package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/afero"

	pb "github.com/wallaceicy06/muni-sign/proto"
)

type SignConfig struct {
	path string
}

func NewSignConfig(path string) *SignConfig {
	return &SignConfig{path: path}
}

var fs afero.Fs = afero.NewOsFs()

func (sc *SignConfig) Get() (*pb.Configuration, error) {
	config, err := readConfigFile(sc.path)
	if err != nil {
		return nil, fmt.Errorf("error getting configuration: %v", err)
	}
	return config, nil
}

func (sc *SignConfig) Put(newConfig *pb.Configuration) error {
	if err := writeConfigFile(sc.path, newConfig); err != nil {
		return fmt.Errorf("error updating configuration: %v", err)
	}
	return nil
}

func readConfigFile(path string) (*pb.Configuration, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening configuration file: %v", err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration: %v", err)
	}
	parsedConfig := &pb.Configuration{}
	if err := proto.UnmarshalText(string(data), parsedConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config proto: %v", err)
	}
	return parsedConfig, nil
}

func writeConfigFile(path string, newConfig *pb.Configuration) error {
	f, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("error opening configuration file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Error closing config file: %v", err)
		}
	}()
	proto.MarshalText(f, newConfig)
	return nil
}
