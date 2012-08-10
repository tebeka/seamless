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

readme:
	rst2html.py README.rst > README.html

.PHONY: all test install fix
