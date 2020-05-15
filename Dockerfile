FROM golang:1.14.2 as build

# Set the Current Working Directory inside the container
WORKDIR /lightchain

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /lightchains/lightpeer ./src/lightpeer/

FROM alpine:latest  

WORKDIR /lightchains/
RUN mkdir ./blockrepo
COPY --from=build /lightchains/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-v", "-repo=${LP_BLOCKREPO}", "-otlp=${LP_OTLPBACKEND}"]