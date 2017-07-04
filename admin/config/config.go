package config

import (
	pb "github.com/wallaceicy06/muni-sign/proto"
)

type SignConfig interface {
	Get() (*pb.Configuration, error)
	Put(*pb.Configuration) error
}
