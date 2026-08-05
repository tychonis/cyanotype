FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN go build -o /cyanotype .

FROM debian:bookworm-slim
COPY --from=build /cyanotype /usr/local/bin/cyanotype
