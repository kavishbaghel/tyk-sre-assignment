FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tyk-sre-assignment .

FROM alpine:latest

RUN addgroup -S tyk-sre-assignment-group && adduser -S tyk-sre-assignment-user -G tyk-sre-assignment-group

WORKDIR /app

COPY --from=builder /app/tyk-sre-assignment .

USER tyk-sre-assignment-user

EXPOSE 8080

ENTRYPOINT ["/app/tyk-sre-assignment"]