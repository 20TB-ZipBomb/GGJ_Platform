FROM golang:1.21.6-alpine3.19 AS build-stage
WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ggjp cmd/server/main.go

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /

COPY --from=build-stage /ggjp /ggjp
EXPOSE 4040

USER nonroot:nonroot

CMD ["/ggjp", "-env", "prod"]