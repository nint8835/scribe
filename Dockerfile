FROM golang:1.24-bookworm AS builder

WORKDIR /build

ARG TAILWIND_VERSION=v3.4.17
RUN curl -L -o /usr/local/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-x64 && \
    chmod +x /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make css && CGO_ENABLED=0 go build

FROM gcr.io/distroless/static AS bot

WORKDIR /bot
COPY --from=builder /build/scribe /bot/scribe

ENTRYPOINT ["/bot/scribe"]
