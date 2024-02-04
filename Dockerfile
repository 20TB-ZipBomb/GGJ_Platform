FROM golang:1.21.6-alpine3.19 AS build-stage
WORKDIR /

COPY . .

RUN go mod download
RUN go build -o ./ggjp cmd/server/main.go

FROM alpine:latest AS build-release-stage
WORKDIR /

COPY --from=build-stage ./ggjp ./
COPY --from=build-stage ./config ./config

ENTRYPOINT [ "./ggjp", "-env", "prod" ]