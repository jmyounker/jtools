CMD = $(shell basename $(CURDIR))

GOFILES = $(wildcard *.go)

clean:
	rm -f $(CMD)

build:
	go build -ldflags "-X main.version=$(VERSION)"

test: build
	@echo do nothing
#	go test

format:
	$(GOFMT) -w $(GOFILES)

install: build test
	install -m 755 -o $(INSTALL_USER) -g $(INSTALL_GROUP) $(CMD) $(PREFIX)/bin/$(CMD)


