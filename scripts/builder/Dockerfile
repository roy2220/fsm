FROM golang:1.13.8-alpine3.11
RUN apk add -q --no-progress --no-cache bash \
                                        make
ADD go.mod go.sum /fsm/
RUN cd /fsm && go mod download
