FROM golang:1.26.0-alpine3.23 AS build
WORKDIR /x
COPY . .
ENV CGO_ENABLED=0
RUN GOBIN=/tmp/bin go install github.com/go-task/task/v3/cmd/task@latest
RUN /tmp/bin/task install
RUN apk --no-cache add upx && \
  upx /x/bin/*

FROM alpine:3.23.3
COPY --from=build /go/bin/ /usr/local/bin/
