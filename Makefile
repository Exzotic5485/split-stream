SERVER_EXECUTABLE=.bin/splitstream-server
CLIENT_EXECUTABLE=.bin/splitstream-client

server: build-server
	.bin/splitstream-server-linux $(filter-out $@,$(MAKECMDGOALS))

client:
	go run cmd/client/client.go $(filter-out $@,$(MAKECMDGOALS))

build: build-server build-client

build-server:
	GOARCH=amd64 GOOS=linux go build -o ${SERVER_EXECUTABLE}-linux ./cmd/server

build-client:
	GOARCH=amd64 GOOS=darwin go build -o ${CLIENT_EXECUTABLE}-darwin cmd/client/client.go
	GOARCH=amd64 GOOS=linux go build -o ${CLIENT_EXECUTABLE}-linux cmd/client/client.go
	GOARCH=amd64 GOOS=windows go build -o ${CLIENT_EXECUTABLE}-windows.exe cmd/client/client.go

clean:
	rm -r .bin
	
%:
    @: