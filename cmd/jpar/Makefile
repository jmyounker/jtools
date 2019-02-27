CMD = $(shell basename $(CURDIR))

TARGET=$(CURDIR)/target

GOFILES = $(wildcard *.go)

clean:
	rm -rf $(TARGET)

update:
	go get

make-target:
	mkdir -p $(TARGET)

build: make-target
	GOBIN=$(TARGET) go install -ldflags "-X main.version=$(VERSION)"

test:
	go test

format:
	$(GOFMT) -w $(GOFILES)

install:
	install -m 755 -o $(INSTALL_USER) -g $(INSTALL_GROUP) $(TARGET)/$(CMD) $(PREFIX)/bin/$(CMD)


