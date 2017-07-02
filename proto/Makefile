all: protoc_muni_sign

clean:
	rm -f *.pb.go
	rm -f *_pb2.py *_pb2_grpc.py

protoc_muni_sign: muni_sign.proto
	protoc -I . --go_out=plugins=grpc:.  muni_sign.proto 
	python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. muni_sign.proto
