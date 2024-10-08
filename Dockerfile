# syntax = docker/dockerfile:1
FROM golang:1.21 AS builder

LABEL authors="chaunsin"

RUN apt-get update && \
    apt-get install -y \
    cmake \
    build-essential \
    curl \
    zip \
    unzip \
    tar

RUN cd /usr/local && git clone https://github.com/google/brotli && \
    mkdir out && cd out && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=./installed /usr/local/brotli && \
    cmake --build . --config Release --target install  && \
    ls -la /usr/local/out/installed/include/ && ls -la /usr/local/out/installed/lib/

RUN cp -r /usr/local/out/installed/lib/* /usr/local/lib/ && \
    cp -r /usr/local/out/installed/include/brotli/ /usr/local/include/

WORKDIR /app
COPY . /app

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod tidy && \
    CGO_ENABLED=1 GOOS=linux go build -o /app/ncmctl cmd/ncmctl/main.go
#    CGO_CFLAGS='-I /usr/local/out/installed/include' \
#    CGO_LDFLAGS='-L /usr/local/out/installed/lib' \
#    LD_LIBRARY_PATH='/usr/local/out/installed/lib' \

FROM frolvlad/alpine-glibc

ENV LD_LIBRARY_PATH=/usr/local/lib:/lib64

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/ncmctl /app
COPY --from=builder /usr/local/out/installed/lib/ /usr/local/lib/
COPY --from=builder /usr/local/out/installed/include/ /usr/local/include/

#CMD ["/app/ncmctl", "-h"]
