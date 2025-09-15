FROM golang:latest AS builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o scanner


FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/scanner ./scanner
ENTRYPOINT ["/app/scanner", "."]
