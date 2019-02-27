all: clean update build test install

.PHONY: all clean

PKG_VERS := github.com/jmyounker/vers

PKG_NAME := jtools

export GOFMT = gofmt -s -tabs=false -tabwidth=4

export COMMANDS := $(shell ls $(CURDIR)/cmd)

export INTERNAL := $(shell ls $(CURDIR)/internal)

clean:
	rm -rf target
	$(foreach cmd,$(COMMANDS),$(MAKE) clean -C cmd/$(cmd);)
	$(foreach module,$(INTERNAL),$(MAKE) clean -C internal/$(module);)

update:
	go get github.com/jmyounker/vers
	$(foreach cmd,$(COMMANDS),$(MAKE) update -C cmd/$(cmd);)
	$(foreach module,$(INTERNAL),$(MAKE) update -C internal/$(module);)

build-vers:
	@echo ------------ $(GOPATH)
	make -C $(GOPATH)/src/$(PKG_VERS) build

set-version: build-vers
	$(eval export VERSION := $(shell $$GOPATH/src/$(PKG_VERS)/vers -f version.json show))
	@echo version $(VERSION)

build: set-version
	$(foreach cmd,$(COMMANDS),$(MAKE) build -C cmd/$(cmd);)
	$(foreach module,$(INTERNAL),$(MAKE) build -C internal/$(module);)

test: build
	$(foreach cmd,$(COMMANDS),$(MAKE) test -C cmd/$(cmd);)
	$(foreach module,$(INTERNAL),$(MAKE) test -C internal/$(module);)

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
	$(eval export INSTALL_USER := $(shell id -u))
endif

set-group:
ifeq ($(shell uname),Darwin)
	$(eval export INSTALL_GROUP := $(shell id -g))
else
	$(eval export INSTALL_GROUP := $(shell id -g))
endif

install: build test set-prefix set-user set-group
	$(foreach cmd,$(COMMANDS),$(MAKE) install -C cmd/$(cmd);)

format:
	$(foreach cmd,$(COMMANDS),$(MAKE) format -C cmd/$(cmd);)
	$(foreach module,$(INTERNAL),$(MAKE) clean -C internal/$(module);)

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package
	$(eval export MODEL_BASE=$(CURDIR)/target/model)

package-osx: set-version set-user set-group package-base
	mkdir $(MODEL_BASE)/osx
	mkdir $(MODEL_BASE)/osx/usr
	mkdir $(MODEL_BASE)/osx/usr/local
	mkdir $(MODEL_BASE)/osx/usr/local/bin
	$(eval export PREFIX=$(MODEL_BASE)/osx/usr/local)
	$(foreach cmd,$(COMMANDS),$(MAKE) install -C cmd/$(cmd);)
	fpm -s dir -t osxpkg -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/osx .

package-rpm: set-version set-user set-group package-base
	mkdir $(MODEL_BASE)/linux-x86-rpm
	mkdir $(MODEL_BASE)/linux-x86-rpm/usr
	mkdir $(MODEL_BASE)/linux-x86-rpm/usr/bin
	$(eval export PREFIX=$(MODEL_BASE)/linux-x86-rpm/usr)
	$(foreach cmd,$(COMMANDS),$(MAKE) install -C cmd/$(cmd);)
	fpm -s dir -t rpm -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-rpm .

package-deb: set-version set-user set-group package-base
	mkdir $(MODEL_BASE)/linux-x86-deb
	mkdir $(MODEL_BASE)/linux-x86-deb/usr
	mkdir $(MODEL_BASE)/linux-x86-deb/usr/bin
	$(eval export PREFIX=$(MODEL_BASE)/linux-x86-deb/usr)
	$(foreach cmd,$(COMMANDS),$(MAKE) install -C cmd/$(cmd);)
	fpm -s dir -t deb -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-deb .
