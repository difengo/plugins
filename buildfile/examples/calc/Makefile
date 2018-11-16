#! /usr/bin/make
#
# Makefile for calc
#
# Targets:
# - "server" builds the micro-service server
# - "client" builds the micro-service client
# - "docker" builds the micro-service docker image 

server:
	go build -a -o ./build/calc ./cmd/calc

client:
	go build -a -o ./build/calc-cli ./cmd/calc-cli

docker:
	docker build -t "wiserskills/calc:1.0.0" .
