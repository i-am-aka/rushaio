FROM debian:stretch-slim

RUN apt update
RUN apt install -y curl unzip
RUN curl -fsSL https://deno.land/x/install/install.sh | sh -s v1.5.4
RUN apt-get remove -y curl

ENV PATH="/root/.deno/bin:${PATH}"

COPY . /app
WORKDIR /app

ENV TZ America/Los_Angeles
ENV PORT=1997
ENV DENO_DIR=./deno
ENTRYPOINT ["deno", "run", "--no-check", "--cached-only", "--unstable", "--allow-all", "serv.ws.js"]

