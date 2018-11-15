#! /usr/bin/make
#
# Makefile for goa v2 Opentracing plugin
#
# Targets:
# - "gen" generates the goa files for the services
# - "example" generates the goa example files

PLUGIN_NAME=opentracing

# include common Makefile content for plugins
include $(GOPATH)/src/goa.design/plugins/plugins.mk

gen:
	@cd examples/calc && \
	goa gen goa.design/plugins/opentracing/examples/calc/design

example:
	@cd examples/calc && \
	goa example goa.design/plugins/opentracing/examples/calc/design