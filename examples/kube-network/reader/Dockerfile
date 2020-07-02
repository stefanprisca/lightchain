FROM golang:1.14.2 as build

# Set the Current Working Directory inside the container
WORKDIR /reader

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/reader .

FROM alpine:latest

WORKDIR /reader/
COPY --from=build /reader/bin .

CMD ["sh", "-c", "./reader -lpAddress=$LP_ADDRESS -persisterAddress=$PERSISTER_ADDRESS"]