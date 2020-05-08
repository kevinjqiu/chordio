.PHONY: vendor

dev:
	goreleaser release --skip-publish --rm-dist --snapshot

build: gen
	cd cmd && go build -o ../bin/chordio .

fmt:
	go fmt ./...

gen:
	protoc -I pb pb/chordio.proto --go_out=plugins=grpc:pb

gen-mock:
	cd chord/node/ && mockery -inpkg -case underscore -name Node
	cd chord/node/ && mockery -inpkg -case underscore -name LocalNode
	cd chord/node/ && mockery -inpkg -case underscore -name RemoteNode
	cd chord/node/ && mockery -inpkg -case underscore -name factory

run:
	docker-compose up -d --service-ports

shell:
	docker-compose run control bash

logs:
	docker-compose logs -f $(n)

join:
	docker-compose run -e CHORDIO_URL=n1:1234 control chordio client join -i n2:2345

test:
	go test -cover ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run-local:
	dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio server -b 127.0.0.1:$(port) -l debug -m 5

vendor:
	# needed temporarily before https://github.com/open-telemetry/opentelemetry-go/issues/682 is fixed
	mkdir -p vendor/
	git clone git@github.com:kevinjqiu/opentelemetry-go.git vendor/opentelemetry-go
	cd vendor/opentelemetry-go && git checkout fix-grpc-method-name-regexp

jaeger:
	docker-compose run --service-ports jaeger

n1:
	dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio server --id 0 -b 127.0.0.1:1234 -l debug -m 3

n1-status:
	CHORDIO_URL=127.0.0.1:1234 dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio client status


n1-join-n2:
	CHORDIO_URL=127.0.0.1:1234 dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio client join -i 127.0.0.1:2345

n2:
	dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio server --id 1 -b 127.0.0.1:2345 -l debug -m 3

n2-status:
	CHORDIO_URL=127.0.0.1:2345 dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio client status

n3:
	dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio server --id 3 -b 127.0.0.1:3456 -l debug -m 3

n3-status:
	CHORDIO_URL=127.0.0.1:3456 dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio client status

n3-join-n1:
	CHORDIO_URL=127.0.0.1:3456 dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio client join -i 127.0.0.1:1234
