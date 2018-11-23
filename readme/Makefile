#! /usr/bin/make
#
# Makefile for goa v2 readme plugin
#
# Targets:
# - "gen" generates the goa files for the example services
# - "example" generates the goa files for the server and client

PLUGIN_NAME=readme
ALIASER_SRC=goa.design/goa/http/dsl

# include common Makefile content for plugins
include $(GOPATH)/src/goa.design/plugins/plugins.mk

gen:
	@cd examples/calc && \
	goa gen goa.design/plugins/readme/examples/calc/design

example:
	@cd examples/calc && \
	goa example goa.design/plugins/readme/examples/calc/design

aliases:
	@echo "Nothing to alias"

test-aliaser: aliases