all: build  

bootstrap:
	go get github.com/siddontang/go-log/log
	go get github.com/siddontang/go-yaml/yaml
	go get github.com/mongodb/mongo-tools

build:
	go install ./...

clean:
	go clean -i ./...

test:
	go test ./...
