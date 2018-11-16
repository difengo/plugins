#! /usr/bin/make
#
# Makefile for goa v2 buildfile plugin
#
# Targets:
# - "gen" generates the goa files for the services
# - "example" generates the example goa files for the services

PLUGIN_NAME=buildfile
ALIASER_SRC=goa.design/goa/http/dsl

# include common Makefile content for plugins
include $(GOPATH)/src/goa.design/plugins/plugins.mk

gen:
	@cd examples/calc && \
	goa gen goa.design/plugins/buildfile/examples/calc/design

example:
	@cd examples/calc && \
	goa example goa.design/plugins/buildfile/examples/calc/design