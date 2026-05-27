FROM golang:1.25.7-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o fireball ./cmd/server

FROM alpine:3.23
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY certs/ /usr/local/share/ca-certificates/
RUN update-ca-certificates
COPY --from=backend-builder /app/fireball .
EXPOSE 8080
CMD ["./fireball"] 

