SHELL := /bin/bash

ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif
ifeq ($(BINDIR),)
    BINDIR := $(PREFIX)/bin
endif
ifeq ($(SYSCONFDIR),)
    SYSCONFDIR := $(PREFIX)/etc
endif
ifeq ($(GO),)
    GO := $(shell type -a -P go | head -n 1)
endif

build:
	mkdir -p build
	cd src; $(GO) build -ldflags "-w -X 'main.systemConfigDir=$(SYSCONFDIR)'" -o ../build/stormfetch stormfetch

install: build/stormfetch config/
	mkdir -p $(DESTDIR)$(BINDIR)
	mkdir -p $(DESTDIR)$(SYSCONFDIR)/stormfetch/
	cp build/stormfetch $(DESTDIR)$(BINDIR)/stormfetch
	cp -r config/. $(DESTDIR)$(SYSCONFDIR)/stormfetch/

compress: build/stormfetch config/
	mkdir -p stormfetch/$(BINDIR)
	mkdir -p stormfetch/$(SYSCONFDIR)/stormfetch/
	cp build/stormfetch stormfetch/$(BINDIR)/stormfetch
	cp -r config/. stormfetch/$(SYSCONFDIR)/stormfetch/
	tar --owner=root --group=root -czf stormfetch.tar.gz stormfetch
	rm -r stormfetch

run: build/stormfetch
	build/stormfetch

clean:
	rm -r build/

.PHONY: build
