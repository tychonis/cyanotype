FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN go build \
    -ldflags="-X 'github.com/tychonis/cyanotype/internal/version.Version=$(git describe --always --dirty)'" \
    -o /cyanotype .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /cyanotype /usr/local/bin/cyanotype
