# syntax=docker/dockerfile:1

FROM golang:1.22.11-alpine

RUN apk add --no-cache git
RUN apk add build-base

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .


RUN go mod download
COPY . .

RUN go build .

EXPOSE 8080
EXPOSE 8081

CMD [ "./retro-chat-rooms" ]