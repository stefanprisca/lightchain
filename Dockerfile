FROM alpine:latest  

WORKDIR /lightchains/
RUN mkdir ./blockrepo
COPY ./bin/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-v", "-repo=${LP_BLOCKREPO}", "-otlp=${LP_OTLPBACKEND}"]