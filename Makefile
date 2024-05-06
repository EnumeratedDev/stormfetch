ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

build:
	mkdir -p build
	go build -o build/stormfetch stormfetch

install: build/stormfetch config/
	mkdir -p $(DESTDIR)$(PREFIX)"/bin/"
	mkdir -p $(DESTDIR)$(PREFIX)"/etc/stormfetch/"
	cp build/stormfetch $(DESTDIR)$(PREFIX)"/bin/stormfetch"
	cp -r config/. $(DESTDIR)$(PREFIX)"/etc/stormfetch/"

clean:
	rm -r build/