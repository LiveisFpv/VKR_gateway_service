FROM golang:1.25-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

# Install CA certificates for outbound HTTPS (SSO validation)
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
