check_install:
	which swagger || go get github.com/go-swagger/go-swagger/cmd/swagger

swagger:
	GO111MODULE=on swagger generate spec -o ./swagger.yml --scan-models