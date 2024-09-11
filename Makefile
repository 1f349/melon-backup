SHELL := /bin/bash
PRODUCT_NAME := melon-backup
BIN := dist/${PRODUCT_NAME}
DNAME := ${PRODUCT_NAME}_
NBIN := ${PRODUCT_NAME}
ENTRY_POINT := ./cmd/${PRODUCT_NAME}
HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
VERSION := ${HASH}
LD_FLAGS := -s -w -X 'main.buildVersion=${VERSION}' -X 'main.buildDate=${BUILD_DATE}' -X 'main.buildName=${PRODUCT_NAME}'
COMP_BIN := /usr/lib/go-1.22/bin/go

ifeq ($(OS),Windows_NT)
	BIN := $(BIN).exe
	DNAME := $(DNAME).exe
	NBIN := $(NBIN).exe
endif

.PHONY: build dev test clean deploy

build: clean
	mkdir -p dist/
	${COMP_BIN} build -o "${BIN}" -ldflags="${LD_FLAGS}" ${ENTRY_POINT}

dev:
	mkdir -p dist/
	${COMP_BIN} build -tags debug -o "${BIN}" -ldflags="${LD_FLAGS}" ${ENTRY_POINT}
	./${BIN}

test:
	${COMP_BIN} test

clean:
	${COMP_BIN} clean
	rm -r -f dist/

deploy: build
	sudo mkdir -p /etc/melon-backup
	sudo cp "${BIN}" /usr/local/bin
	sudo "/usr/local/bin/$(NBIN)" generate -config=/etc/melon-backup/example.yml
