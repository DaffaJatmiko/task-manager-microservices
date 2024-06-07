# build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY /loggerServiceApp /app

CMD ["/app/loggerServiceApp"]