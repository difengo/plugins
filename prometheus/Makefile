#! /usr/bin/make
#
# Makefile for goa v2 prometheus plugin
#
# Targets:
# - "gen" generates the goa files for the example services

PLUGIN_NAME=prometheus
ALIASER_SRC=goa.design/goa/http/dsl

# include common Makefile content for plugins
include $(GOPATH)/src/goa.design/plugins/plugins.mk

gen:
	@cd examples/calc && \
	goa gen goa.design/plugins/prometheus/examples/calc/design

example:
	@cd examples/calc && \
	goa example goa.design/plugins/prometheus/examples/calc/design

aliases:
	@echo "Nothing to alias"

test-aliaser: aliases
