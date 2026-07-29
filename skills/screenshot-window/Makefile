BINARY := bin/wincap

.PHONY: build clean

build:
	swift build -c release
	mkdir -p bin
	cp .build/release/wincap $(BINARY)

clean:
	rm -rf .build $(BINARY)
