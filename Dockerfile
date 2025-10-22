# syntax=docker/dockerfile:1
FROM diningclub/golang-dev-tools:latest

WORKDIR /src/app

COPY go.mod go.sum /src/app/
RUN go mod download

COPY . /src/app
