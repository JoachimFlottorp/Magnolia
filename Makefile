run:
	go build -o bin/main cmd/server/main.go
	./bin/main

debug:
	go build -o bin/main cmd/server/main.go
	./bin/main -debug

build:
	go build -o bin/main cmd/server/main.go

docs:
	swagger generate spec -m -o ./web/public/swagger.json

proto:
	cd protobuf; npm run generate

docker:
	docker build -f docker/Dockerfile.shared -t jf/magnolia.deps .

compose: docker
	docker compose build
	docker compose up

test:
	go test -v ./...

coverage:
	go test -race -v -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
