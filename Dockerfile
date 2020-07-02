FROM golang:1.14.4 as build

# Set the Current Working Directory inside the container
WORKDIR /lightpeer

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/lightpeer ./src/lightpeer/

FROM alpine:latest

WORKDIR /lightpeer/
RUN mkdir ./blockrepo
COPY --from=build /lightpeer/bin/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-repo=./blockrepo"]