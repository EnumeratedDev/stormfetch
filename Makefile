ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif
ifeq ($(BINDIR),)
    BINDIR := $(PREFIX)/bin
endif
ifeq ($(SYSCONFDIR),)
    SYSCONFDIR := $(PREFIX)/etc
endif

build:
	mkdir -p build
	go build -o build/stormfetch stormfetch

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

clean:
	rm -r build/