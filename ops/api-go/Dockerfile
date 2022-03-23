FROM golang:1.14.4-buster

RUN apt update
RUN apt install -y unzip htop vim lsof

RUN mkdir /out

WORKDIR /go
RUN mkdir -p src/golang.org/x/
RUN mkdir -p src/github.com/refraction-networking/
RUN mkdir -p src/github.com/elazarl/goproxy
COPY github.com/refraction-networking/utls src/github.com/refraction-networking/utls
COPY github.com/elazarl/goproxy src/github.com/elazarl/goproxy
COPY golang.org/x/net src/golang.org/x/net
RUN mkdir src/rush/
WORKDIR /go/src/rush

COPY rush/go.mod .
COPY rush/go.sum .
COPY rush/api api
COPY rush/akgen akgen
COPY rush/common common

WORKDIR /go/src/rush
RUN GO111MODULE=on go get github.com/minio/minio-go/v7
RUN GO111MODULE=on go get -d -v ./...
WORKDIR /go/src/rush/api

RUN go install

CMD ["./copy.sh"]
