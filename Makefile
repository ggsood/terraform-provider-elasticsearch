SHELL := /bin/bash
export GO111MODULE ?= on
export VERSION := 0.0.1
export BINARY := terraform-provider-elasticsearch
export GOBIN = $(shell pwd)/bin

include scripts/Makefile.help
.DEFAULT_GOAL := help

include build/Makefile.build
include build/Makefile.test
include build/Makefile.deps
include build/Makefile.tools
include build/Makefile.lint
include build/Makefile.format
include build/Makefile.release