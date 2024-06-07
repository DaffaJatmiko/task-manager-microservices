FROM alpine:latest

RUN mkdir /app

COPY /taskApp /app

CMD ["/app/taskApp"]