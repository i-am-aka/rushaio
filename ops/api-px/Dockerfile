FROM golang:1.14.4-buster


RUN apt update
RUN apt install -y unzip htop vim lsof
RUN mkdir -p /root/.deno/bin
COPY rush/external/deno_linux /root/.deno/bin/deno
ENV PATH="/root/.deno/bin:${PATH}"

WORKDIR /go
RUN mkdir src/rush/

COPY rush/pxgen src/rush/pxgen

RUN go get -d -v ./...

ENV TZ America/Los_Angeles
WORKDIR /go/src/rush/pxgen/api
ENV PORT=8082
CMD ["go", "run", "."]

