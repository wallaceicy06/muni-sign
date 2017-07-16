all: build
build: muni_sign.proto 
	python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. --go_out=plugins=grpc:. muni_sign.proto
clean:
	rm -f *.pb.go
	rm -f *_pb2.py *_pb2_grpc.py
test:
	true
