# Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
#
# Use of this source code is governed by the license in the LICENSE file
# included with the source code.
#
# The Makefile is not required, a 'go build -o bin/ ./...' works as well :)
#
# Makefile to build WolfMUD. Targets of note:
#
#   build       - native build (default)
#   build-all   - build for all supported platforms
#   build-race  - build with race detector
#   run         - start server with logging to terminal and bin/log
#   batch       - start server with logging to bin/log only
#   race        - start server with race detector enabled, logging to terminal and bin/log
#   test        - run tests
#   cover       - run tests with coverage collection and display in browser
#   doc         - start godoc server with notes turned on
#   clean       - clean bin directory
#
SHELL := /bin/bash

export CGO_ENABLED=0
export GOPROXY=off
export GORACE=history_size=7 halt_on_error=1
export GOCACHE=/tmp/go-build
export TZ=Europe/London
export WOLFMUD_DIR=../data

VERSION := $(shell git describe --dirty)
LDFLAGS := -ldflags "-X code.wolfmud.org/WolfMUD.git/cmd.commit=$(VERSION)"

# Standard native build
build: version
	go build -o bin/ -v $(LDFLAGS) -gcflags="-e" ./...

build-all: build linux-amd64 linux-386 linux-arm5 linux-arm6 linux-arm7 windows-amd64 windows-386

# Build targets also used by release/Makefile
linux-amd64: version
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/ --trimpath -v $(LDFLAGS) ./...
linux-386: version
	GOOS=linux GOARCH=386 go build -o bin/linux-386/ --trimpath -v $(LDFLAGS) ./...
linux-arm5: version
	GOOS=linux GOARCH=arm GOARM=5 go build -o bin/linux-arm5/ --trimpath -v $(LDFLAGS) ./...
linux-arm6: version
	GOOS=linux GOARCH=arm GOARM=6 go build -o bin/linux-arm6/ --trimpath -v $(LDFLAGS) ./...
linux-arm7: version
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/linux-arm7/ --trimpath -v $(LDFLAGS) ./...
windows-amd64: version
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/ --trimpath -v $(LDFLAGS) ./...
windows-386: version
	GOOS=windows GOARCH=386 go build -o bin/windows-386/ --trimpath -v $(LDFLAGS) ./...

# Run server with logging to terminal and bin/log
run: build bin/log
	cd bin ;\
	clear ;\
	./server 2>&1 | tee log/`date -u +%F-%T`.log

# Run server with logging to bin/log only
batch: build bin/log
	cd bin ;\
	clear ;\
	./server > log/`date -u +%F-%T`.log 2>&1

# Build with race detector
build-race: version
	CGO_ENABLED=1 go build -o ./bin/ -race -trimpath -v $(LDFLAGS) -gcflags -e ./...

# Run with race detector and logging to terminal and bin/log
race: build-race bin/log
	cd bin ;\
	clear ;\
	./server 2>&1 | tee log/`date -u +%F-%T`.log

bin/log:
	mkdir -p bin/log

# If our git commit ID changes touch version.go so that we relink and detect the
# change even if no source files have actually changed
version: build/version.txt
ifneq "$(shell cat build/version.txt)" "$(VERSION)"
	echo $(VERSION) > build/version.txt ;\
	touch ./cmd/version.go;
endif

build/version.txt:
	touch build/version.txt

# Display Go environment as seen from the makefile for debugging
env:
	go env

vet:
	go vet ./...

.PHONY: test
test:
	WOLFMUD_DIR="NONE" go test -cover ./...

.PHONY: race-test
race-test:
	WOLFMUD_DIR="NONE" CGO_ENABLED=1 go test -race -cover ./...

.PHONY: cover
cover:
	WOLFMUD_DIR="NONE" go test -coverprofile bin/cover.out ./...; \
	go tool cover -html=bin/cover.out

.PHONY: bench
bench:
	WOLFMUD_DIR="NONE" go test -run NONE -timeout 10m -bench "." -benchtime 1s ./...

.PHONY: longbench
longbench:
	WOLFMUD_DIR="NONE" go test -run NONE -timeout 10m -bench "." -benchtime 10s ./...

.PHONY: doc
doc:
	cd $(GOPATH)/src/ ;\
	godoc -v -http=:6060 -notes="BUG|TODO|FIXME"

.PHONY: clean
clean:
	find bin -type f -executable -delete ;\
	find bin/log/ -type f -delete ;\
	find bin -name "*prof" -delete ;\
	find bin -name "cover.out" -delete ;\
	go clean -cache -modcache
