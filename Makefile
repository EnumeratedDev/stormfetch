# Installation paths
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
LIBEXECDIR ?= $(PREFIX)/libexec
SYSCONFDIR := $(PREFIX)/etc

# Compilers and tools
GO ?= go

# Compiler flags
GOOS ?= $(shell $(GO) env | grep '^GOOS' | cut -d'=' -f2 | tr -d "'")
GOARCH ?= $(shell $(GO) env | grep '^GOARCH' | cut -d'=' -f2 | tr -d "'")
LDFLAGS ?= -w

# Build-time variables
VERSION ?= $(shell git describe --tags --dirty)

build: build-stormfetch build-stormfetch-monitor-detection

build-stormfetch:
	install -d build
	cd src/stormfetch; GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS) -X 'main.StormfetchVersion=$(VERSION)' -X 'main.SystemConfigDir=$(SYSCONFDIR)' -X 'main.LibexecDir=$(LIBEXECDIR)'" -o ../../build/stormfetch

build-stormfetch-monitor-detection:
	install -d build
	cd src/stormfetch-monitor-detection; GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS)" -o ../../build/stormfetch-monitor-detection

install: install-stormfetch install-stormfetch-monitor-detection install-config

install-stormfetch: build/stormfetch
	# Create directory
	install -dm755 $(DESTDIR)$(BINDIR)
	#  Install binary
	install -m755 build/stormfetch $(DESTDIR)$(BINDIR)/stormfetch

install-stormfetch-monitor-detection: build/stormfetch-monitor-detection
	# Create directory
	install -dm755 $(DESTDIR)$(LIBEXECDIR)
	#  Install binary
	install -m755 build/stormfetch-monitor-detection $(DESTDIR)$(LIBEXECDIR)/stormfetch-monitor-detection

install-config:
	# Create directory
	install -dm755 $(DESTDIR)$(SYSCONFDIR)
	# Install files
	cp -r config $(DESTDIR)$(SYSCONFDIR)/stormfetch/

uninstall:
	-rm -f $(DESTDIR)$(BINDIR)/stormfetch
	-rm -f $(DESTDIR)$(LIBEXEDIR)/stormfetch-monitor-detection
	-rm -rf $(DESTDIR)$(SYSCONFDIR)/stormfetch

clean:
	rm -r build/

.PHONY: build build-stormfetch build-stormfetch-monitor-detection install install-config uninstall clean
