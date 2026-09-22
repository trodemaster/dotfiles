.PHONY: install build test clean

install:
	GOWORK=off go install .

build:
	GOWORK=off go build -o copilot-statusline .

test:
	GOWORK=off go test ./...

clean:
	rm -f copilot-statusline
