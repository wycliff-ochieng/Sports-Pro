FROM golang:1.21-alpine as builder

#install dependancies
RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

#set work dir
WORKDIR /app


COPY go.mod go.sum ./

#download dependancies
RUN go mod download


#copy source code into container
#COPY ./auth-service ./
COPY . . 

#build the go application
RUN go build -o authentication-service

#run binary
ENTRYPOINT["./authentication-service"]