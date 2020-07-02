FROM golang:1.14.2 as build

# Set the Current Working Directory inside the container
WORKDIR /persister

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/persister .

FROM alpine:latest

WORKDIR /persister/
COPY --from=build /persister/bin .

CMD ["sh", "-c", "./persister -lpAddress=$LP_ADDRESS"]