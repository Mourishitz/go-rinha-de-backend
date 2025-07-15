FROM alpine:latest

RUN mkdir /app

COPY builds/go-proxy/goProxy /app

CMD ["/app/goProxy"]
