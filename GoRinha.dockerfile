FROM alpine:latest

RUN mkdir /app

COPY dist/rinha /app

CMD ["/app/rinha"]

