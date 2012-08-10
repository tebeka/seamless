export GOPATH := $(shell dirname $(shell dirname $(PWD)))
PACKAGE := seesaw

all:
	go build $(PACKAGE)

test:
	go test -v $(PACKAGE)

fix:
	go fix $(PACKAGE)

install:
	go install $(PACKAGE)

.PHONY: all test install fix
