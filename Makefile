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

compose:
	docker build -f Dockerfile.shared -t jf/yeahapi.deps .
	docker compose up