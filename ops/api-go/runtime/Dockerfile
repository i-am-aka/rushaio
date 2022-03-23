FROM debian:stretch-slim

RUN apt update
RUN apt install -y curl unzip
RUN curl -fsSL https://deno.land/x/install/install.sh | sh -s v1.5.4
RUN apt-get remove -y curl

ENV PATH="/root/.deno/bin:${PATH}"

RUN mkdir -p /api/bin
RUN mkdir -p /api/akgen/fp/sensor

COPY out/adyen_fp.json /api/bin
COPY out/mact.gob.gz /api/akgen/fp
COPY out/sensors.gob.gz /api/akgen/fp/sensor
COPY out/api /api/bin
COPY out/js /api/akgen/js

WORKDIR /api/bin

ENV PORT=8192
CMD ["./api"]
