builds: build build-linux

build:
	go build -mod=vendor -o ./bin/egamifi
build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/egamifi-linux
deploy:
	scp ./bin/egamifi-linux sakura2:~/bin/egamifi-linux

