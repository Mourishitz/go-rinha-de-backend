FROM alpine:latest

RUN mkdir /app

COPY builds/go-worker/goWorker /app

CMD ["/app/goWorker"]
