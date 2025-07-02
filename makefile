build:
	go build -o bin/app cmd/main.go

run: build
	./bin/app

clean:
	rm -rf bin/*

test:
	go test -v ./... -count=1

integration-test:
	docker-compose up --build integration-tests

docker-up:
	docker-compose up --build

.PHONY: build run clean test