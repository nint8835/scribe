FROM golang:1.16-buster

WORKDIR /build
COPY . .
RUN go mod download &&\
    go mod verify &&\
    go build

ENTRYPOINT ["/build/scribe"]