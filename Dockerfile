FROM golang:1.24-alpine AS builder 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apk add --no-cache openssl-dev gcc musl-dev
RUN apk add --no-cache openssl-dev
ENV PKG_CONFIG_PATH="/usr/lib/pkgconfig:/usr/share/pkgconfig"

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /bot ./cmd/main.go


FROM alpine:3.18

WORKDIR /app

COPY --from=builder /bot /app/bot
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/init.lua /app/init.lua

RUN apk --no-cache add ca-certificates

CMD ["/app/bot"]