GPATH=./gen/fileservice/

all:
	@protoc -I proto proto/tages.proto \
	--go_out=$(GPATH) --go-grpc_out=$(GPATH) \
	--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative
	@echo "Files generated"

cleangen:
	@rm -rf $(GPATH)*.go

clean:
	@rm -rf ./storage/files