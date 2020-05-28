FROM alpine:latest  

WORKDIR /lightchains/
RUN mkdir ./blockrepo
COPY ./bin/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-repo=./blockrepo"]