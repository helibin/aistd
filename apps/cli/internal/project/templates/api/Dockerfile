FROM golang:1.23.12-alpine3.20 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/server
FROM alpine:3.20.3
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /out/api /api
USER app
EXPOSE 8080
ENTRYPOINT ["/api"]
