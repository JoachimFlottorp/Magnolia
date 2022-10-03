run:
	go build -o bin/main cmd/server/main.go
	./bin/main

debug:
	go build -o bin/main cmd/server/main.go
	./bin/main -debug

build:
	go build -o bin/main cmd/server/main.go

proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		protobuf/collector/collector.proto

docs:
	swagger generate spec -m -o ./web/public/swagger.json
