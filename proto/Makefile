all: build
build: muni_sign.proto 
	protoc -I . --plugin=protoc-gen-grpc_python=$(shell which grpc_python_plugin) --go_out=plugins=grpc:. --python_out=. --grpc_python_out=. muni_sign.proto 
clean:
	rm -f *.pb.go
	rm -f *_pb2.py *_pb2_grpc.py

