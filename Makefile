SERVER_EXECUTABLE=.bin/splitstream-server
CLIENT_EXECUTABLE=.bin/splitstream-client

server:
	go run cmd/server/server.go $(filter-out $@,$(MAKECMDGOALS))

client:
	go run cmd/client/client.go $(filter-out $@,$(MAKECMDGOALS))

build: build-server build-client

build-server:
	GOARCH=amd64 GOOS=darwin go build -o ${SERVER_EXECUTABLE}-darwin cmd/server/server.go
	GOARCH=amd64 GOOS=linux go build -o ${SERVER_EXECUTABLE}-linux cmd/server/server.go
	GOARCH=amd64 GOOS=windows go build -o ${SERVER_EXECUTABLE}-windows.exe cmd/server/server.go

build-client:
	GOARCH=amd64 GOOS=darwin go build -o ${CLIENT_EXECUTABLE}-darwin cmd/client/client.go
	GOARCH=amd64 GOOS=linux go build -o ${CLIENT_EXECUTABLE}-linux cmd/client/client.go
	GOARCH=amd64 GOOS=windows go build -o ${CLIENT_EXECUTABLE}-windows.exe cmd/client/client.go

clean:
	rm -r .bin
	
%:
    @: