GOFMT=gofmt -s -tabs=false -tabwidth=4

GOFILES=mustache.go

clean:

update:
	go get

build:
	go build

test:
	go test

format:
	${GOFMT} -w ${GOFILES}
	${GOFMT} -w mustache_test.go 
