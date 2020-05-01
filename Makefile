build: gen
	cd cmd && go build -o ../bin/chordio .

gen:
	protoc -I pb pb/chordio.proto --go_out=plugins=grpc:pb
