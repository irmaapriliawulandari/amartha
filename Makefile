.PHONY: run cron mq build test clean

run:
	go run .

cron:
	go run ./cron

mq:
	go run ./mq

build:
	go build -o bin/amartha-test .
	go build -o bin/cron ./cron
	go build -o bin/mq ./mq

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

clean:
	rm -rf bin coverage.out
