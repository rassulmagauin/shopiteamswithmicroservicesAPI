check_protos:
	which protoc-gen-go || go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


protos: check_protos
	protoc -I protos/ protos/*.proto --go_out=protos/currency --go-grpc_out=protos/currency

