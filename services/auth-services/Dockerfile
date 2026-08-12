FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /app 

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o main-app cmd/app/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/main-app .

EXPOSE 9091

ENTRYPOINT ["./main-app"]