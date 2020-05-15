FROM golang:1.14.2 as build

# Set the Current Working Directory inside the container
WORKDIR /lightpeer-client

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/client-otel ./examples/client/

FROM alpine:latest

WORKDIR /lightpeer-client/
COPY --from=build /lightpeer-client/bin .

CMD ["sh", "-c", "./client-otel -lpAddress=$LP_ADDRESS"]