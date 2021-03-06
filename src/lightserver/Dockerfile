#// Copyright 2020 Stefan Prisca
# //
# // Licensed under the Apache License, Version 2.0 (the "License");
# // you may not use this file except in compliance with the License.
# // You may obtain a copy of the License at
# //
# //     http://www.apache.org/licenses/LICENSE-2.0
# //
# // Unless required by applicable law or agreed to in writing, software
# // distributed under the License is distributed on an "AS IS" BASIS,
# // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# // See the License for the specific language governing permissions and
# // limitations under the License.

FROM golang:1.15 as build

# Set the Current Working Directory inside the container
WORKDIR /lightpeer

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/lightpeer .

FROM alpine:latest

WORKDIR /lightpeer/
RUN mkdir ./blockrepo
COPY --from=build /lightpeer/bin/lightpeer .

EXPOSE 9081

CMD ["./lightpeer", "-repo=./blockrepo"]