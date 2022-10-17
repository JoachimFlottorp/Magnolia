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
	export PYTHONPATH
	cd protobuf; python3 generate.py

compose:
	docker build -f Dockerfile.shared -t jf/yeahapi.deps .
	docker compose build
	docker compose up

test:
	go test -v ./...

coverage:
	go test -race -v -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
