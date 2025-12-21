# Installation paths
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR := $(PREFIX)/etc

# Compilers and tools
GO ?= go

# Build-time variables
VERSION ?= $(shell git describe --tags --dirty)

build:
	mkdir -p build
	cd src; $(GO) build -ldflags "-w -X 'main.StormfetchVersion=$(VERSION)' -X 'main.SystemConfigDir=$(SYSCONFDIR)'" -o ../build/stormfetch stormfetch

install: build/stormfetch config/
	# Create directoriy
	install -dm755 $(DESTDIR)$(BINDIR)
	#  Install binary
	install -m755 build/stormfetch $(DESTDIR)$(BINDIR)/stormfetch

install-config:
	# Create directory
	install -dm755 $(DESTDIR)$(SYSCONFDIR)
	# Install files
	cp -r config $(DESTDIR)$(SYSCONFDIR)/stormfetch/

uninstall:
	-rm -f $(DESTDIR)$(BINDIR)/stormfetch
	-rm -rf $(DESTDIR)$(SYSCONFDIR)/stormfetch

clean:
	rm -r build/

.PHONY: build install install-config uninstall clean
