ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o /app/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 9090
CMD ["/app/server"]

