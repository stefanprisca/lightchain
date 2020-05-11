FROM golang:1.14.2-alpine3.11 as build

# Set the Current Working Directory inside the container
WORKDIR lightchain

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /lightchains/lightpeer ./src/lightpeer/

FROM alpine:latest  

WORKDIR /lightchains/
RUN mkdir ./blockrepo
COPY --from=build /lightchains/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-repo=$LP_BLOCKREPO"]