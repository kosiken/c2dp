FROM rust:1-bookworm AS signing-tool
ARG C2PATOOL_VERSION=0.27.22
RUN cargo install c2patool --version ${C2PATOOL_VERSION} --locked

FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/go-api ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libssl3 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app && useradd --system --gid app app
WORKDIR /app
COPY --from=build /bin/go-api /app/go-api
COPY --from=signing-tool /usr/local/cargo/bin/c2patool /usr/local/bin/c2patool
ENV C2PATOOL_PATH=/usr/local/bin/c2patool
RUN mkdir -p /app/uploads && chown -R app:app /app
USER app
EXPOSE 8080
CMD ["/app/go-api"]
