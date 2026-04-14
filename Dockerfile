FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /stealthfetch .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /stealthfetch /usr/local/bin/stealthfetch
EXPOSE 8899
ENTRYPOINT ["stealthfetch"]
