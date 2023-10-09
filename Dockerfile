FROM golang:1.21 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/bot/main.go

FROM alpine:latest AS production
# Add this line to install timezone data
RUN apk add --no-cache tzdata

COPY --from=builder /app .
CMD ["./app"]