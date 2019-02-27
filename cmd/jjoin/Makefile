CMD = $(shell basename $(CURDIR))

GOFILES = $(wildcard *.go)

clean:
	rm -f $(CMD)

update:
	go get

build:
	go build -ldflags "-X main.version=$(VERSION)"

test:
	go test

format:
	$(GOFMT) -w $(GOFILES)

install:
	install -m 755 -o $(INSTALL_USER) -g $(INSTALL_GROUP) $(CMD) $(PREFIX)/bin/$(CMD)



