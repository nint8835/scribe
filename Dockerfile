FROM golang:1.24 AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y libsqlite3-dev

ARG TAILWIND_VERSION=v4.0.6
RUN wget -O /usr/local/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-x64 && \
    chmod +x /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make prod-css && CGO_ENABLED=1 go build -a -ldflags '-linkmode external -extldflags "-static"' --tags fts5

FROM gcr.io/distroless/static AS bot

WORKDIR /bot
COPY --from=builder /build/scribe /bot/scribe

ENTRYPOINT ["/bot/scribe"]
