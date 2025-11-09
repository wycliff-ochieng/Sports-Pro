FROM golang:1.21-alpine as builder


RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*


WORKDIR /app

#copy dependancies and packages 
COPY go.mod go.sum ./

#copy source code into container 
COPY . .

RUN go build -o team-service


ENTRYPOINT["./team-service"]