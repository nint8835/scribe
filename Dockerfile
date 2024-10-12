FROM golang:1.23-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static AS bot

WORKDIR /bot
COPY --from=builder /build/scribe /bot/scribe

ENTRYPOINT ["/bot/scribe"]
