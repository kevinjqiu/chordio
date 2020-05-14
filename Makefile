.PHONY: vendor

CMD=dist/chordio_$$(uname | tr '[:upper:]' '[:lower:]')_amd64/chordio

dev:
	goreleaser release --skip-publish --rm-dist --snapshot

build: gen
	cd cmd && go build -o ../bin/chordio .

fmt:
	go fmt ./...

vet:
	go vet ./...

gen:
	protoc -I pb pb/chordio.proto --go_out=plugins=grpc:pb

gen-mock:
	cd chord/node/ && mockery -inpkg -case underscore -testonly -name Node
	cd chord/node/ && mockery -inpkg -case underscore -testonly -name LocalNode
	cd chord/node/ && mockery -inpkg -case underscore -testonly -name RemoteNode
	cd chord/node/ && mockery -inpkg -case underscore -testonly -name factory

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
	$(CMD) server -b 127.0.0.1:$(port) -m 5

vendor:
	# needed temporarily before https://github.com/open-telemetry/opentelemetry-go/issues/682 is fixed
	mkdir -p gomod/
	git clone https://github.com/kevinjqiu/opentelemetry-go.git gomod/opentelemetry-go
	cd gomod/opentelemetry-go && git checkout fix-grpc-method-name-regexp

jaeger:
	docker-compose run --service-ports jaeger

n1:
	$(CMD) server --id 0 -b 127.0.0.1:1234  -m 3

n1-status:
	CHORDIO_URL=127.0.0.1:1234 $(CMD) client status

n1-stabilize:
	CHORDIO_URL=127.0.0.1:1234 $(CMD) client stabilize

n1-fixfingers:
	CHORDIO_URL=127.0.0.1:1234 $(CMD) client fixfingers

n1-join-n2:
	CHORDIO_URL=127.0.0.1:1234 $(CMD) client join -i 127.0.0.1:2345

n2:
	$(CMD) server --id 1 -b 127.0.0.1:2345 -m 3

n2-status:
	CHORDIO_URL=127.0.0.1:2345 $(CMD) client status

n2-stabilize:
	CHORDIO_URL=127.0.0.1:2345 $(CMD) client stabilize

n2-fixfingers:
	CHORDIO_URL=127.0.0.1:2345 $(CMD) client fixfingers

n3:
	$(CMD) server --id 3 -b 127.0.0.1:3456 -m 3

n3-status:
	CHORDIO_URL=127.0.0.1:3456 $(CMD) client status

n3-join-n1:
	CHORDIO_URL=127.0.0.1:3456 $(CMD) client join -i 127.0.0.1:1234

n3-join-n2:
	CHORDIO_URL=127.0.0.1:3456 $(CMD) client join -i 127.0.0.1:2345

test-join: n1-join-n2 n3-join-n1 n1-status n2-status n3-status
