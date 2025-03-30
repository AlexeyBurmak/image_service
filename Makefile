GPATH=./gen/fileservice/

installproto:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

genproto:
	@protoc -I proto proto/tages.proto \
	--go_out=$(GPATH) --go-grpc_out=$(GPATH) \
	--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative
	@echo "Files generated"

runserver:
	@go run main.go

runclient:
	@go run ./client/main.go

cleangen:
	@rm -rf $(GPATH)*.go
	@echo "Files deleted"

clean:
	@rm -rf ./storage/* ./client/download/*