dev:
	goreleaser release --skip-publish --rm-dist --snapshot

build: gen
	cd cmd && go build -o ../bin/chordio .

gen:
	protoc -I pb pb/chordio.proto --go_out=plugins=grpc:pb

run:
	docker-compose up -d
	docker-compose exec control bash

test:
	go test -v ./...
