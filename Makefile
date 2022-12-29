BINS = $(shell ls -1 cmd)

run: build
	./bin/server

debug: build
	./bin/server -debug

build:
	@for bin in $(BINS); do \
		printf "Building $$bin...\n"; \
		go build -o bin/$$bin cmd/$$bin/main.go; \
	done

docs:
	swagger generate spec -m -o ./web/public/swagger.json

proto:
	make -C protobuf generate

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
