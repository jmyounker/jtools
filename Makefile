all: clean update build test install

.PHONY: all clean

PKG_VERS := github.com/jmyounker/vers

PKG_NAME := jtools

export GOFMT=gofmt -s

clean:
	rm -rf target
	$(MAKE) clean -C cmd/jc
	$(MAKE) clean -C cmd/jjoin
	$(MAKE) clean -C cmd/jpar
	$(MAKE) clean -C cmd/jx

update:
	go get $(PKG_VERS)

build-vers:
	make -C $$GOPATH/src/$(PKG_VERS) build

set-version: build-vers
	$(eval export VERSION := $(shell $$GOPATH/src/$(PKG_VERS)/vers -f version.json show))
	@echo version $(VERSION)

build: set-version
	$(MAKE) build -C cmd/jc
	$(MAKE) build -C cmd/jjoin
	$(MAKE) build -C cmd/jpar
	$(MAKE) build -C cmd/jx

test: build
	$(MAKE) test -C cmd/jc
	$(MAKE) test -C cmd/jjoin
	$(MAKE) test -C cmd/jpar
	$(MAKE) test -C cmd/jx

set-prefix:
ifndef PREFIX
ifeq ($(shell uname),Darwin)
	$(eval export PREFIX := /usr/local)
else
	$(eval export PREFIX := /usr)
endif
endif

set-user:
ifeq ($(shell uname),Darwin)
	$(eval export INSTALL_USER := $(shell id -u))
else
	$(eval export INSTALL_USER := root)
endif

set-group:
ifeq ($(shell uname),Darwin)
	$(eval export INSTALL_GROUP := $(shell id -g))
else
	$(eval export INTALL_GROUP := root)
endif

install: build test set-prefix set-user set-group
	$(MAKE) -C cmd/jc install
	$(MAKE) -C cmd/jpar install

format:
	$(MAKE) jc -C cmd/jc

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package

package-osx: set-version package-base
	mkdir target/model/osx
	mkdir target/model/osx/usr
	mkdir target/model/osx/usr/local
	mkdir target/model/osx/usr/local/bin
	install -m 755 $(CMD) target/model/osx/usr/local/bin/$(CMD)
	fpm -s dir -t osxpkg -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/osx .

package-rpm: set-version package-base
	mkdir target/model/linux-x86-rpm
	mkdir target/model/linux-x86-rpm/usr
	mkdir target/model/linux-x86-rpm/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-rpm/usr/bin/$(CMD)
	fpm -s dir -t rpm -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-rpm .

package-deb: set-version package-base
	mkdir target/model/linux-x86-deb
	mkdir target/model/linux-x86-deb/usr
	mkdir target/model/linux-x86-deb/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-deb/usr/bin/$(CMD)
	fpm -s dir -t deb -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-deb .

