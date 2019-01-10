FROM golang:alpine as builder
LABEL maintainer="juanjcsr@gmail.com"

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN apk add --no-cache \
        bash 
COPY . /go/src/github.com/juanjcsr/mpupldr

RUN apk add --no-cache \
        git \ 
        gcc \
        g++ \
        libc-dev \
        sqlite-dev \
        libgcc \
        zlib-dev \
        make 

RUN cd /go/src/github.com/juanjcsr/mpupldr \
        && GO111MODULE=on go build

FROM alpine:latest

RUN apk add --no-cache \
        bash \
        git \ 
        gcc \
        g++ \
        libc-dev \
        sqlite-dev \
        libgcc \
        zlib-dev \
        make 

RUN git clone https://github.com/mapbox/tippecanoe.git && cd tippecanoe && make -j && make install

COPY --from=builder /go/src/github.com/juanjcsr/mpupldr/mpupldr /usr/bin/mpupldr


ENTRYPOINT ["mpupldr"]
CMD ["--help"]