export GOPATH := $(shell dirname $(shell dirname $(PWD)))
PACKAGE := seamless

all:
	go build $(PACKAGE)

test:
	go test -v $(PACKAGE)

fix:
	go fix $(PACKAGE)

install:
	go install $(PACKAGE)

.PHONY: all test install fix
