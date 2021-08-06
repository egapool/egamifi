build:
	go build -mod=vendor -o ./bin/egamifi
build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/egamifi-linux

