FROM golang:1.16-bullseye

WORKDIR /build
COPY . .
RUN go mod download &&\
    go mod verify &&\
    go build

ENTRYPOINT ["/build/scribe"]
