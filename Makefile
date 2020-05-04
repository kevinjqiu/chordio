dev:
	goreleaser release --skip-publish --rm-dist --snapshot

build: gen
	cd cmd && go build -o ../bin/chordio .

gen:
	protoc -I pb pb/chordio.proto --go_out=plugins=grpc:pb

run:
	docker-compose up -d

shell:
	docker-compose run control bash

logs:
	docker-compose logs -f $(n)

join:
	docker-compose run -e CHORDIO_URL=n1:1234 control chordio client join -i n2:2345

test:
	go test -v -cover ./...

run-local:
	dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio server -b 127.0.0.1:$(port) -l debug -m 5

vendor:
	# needed temporarily before https://github.com/open-telemetry/opentelemetry-go/issues/682 is fixed
	mkdir vendor/
	git clone git@github.com:kevinjqiu/opentelemetry-go.git vendor/
	cd vendor/opentelemetry-go && git checkout fix-grpc-method-name-regexp
