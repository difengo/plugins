# Dockerfile for calc micro-service
FROM golang:1.10.3-alpine AS builder

RUN apk add --no-cache git && \
    go version

COPY . /go/src/wiserskills/calc

WORKDIR /go/src/wiserskills/calc

RUN go get github.com/golang/dep/cmd/dep && \
    dep init && \
    dep ensure -v

WORKDIR /go/src/wiserskills/calc/cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./calc .

FROM scratch

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.name="calc"
LABEL org.label-schema.description="HTTP service for adding numbers, a goa teaser"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.vendor="wiserskills"
LABEL org.label-schema.version="1.0.0"


WORKDIR /root/

COPY --from=builder /go/src/wiserskills/calc/cmd/calc .

EXPOSE 80 8080
ENTRYPOINT ["./calc"]
