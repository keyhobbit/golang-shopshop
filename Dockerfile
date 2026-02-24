FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/admin ./cmd/admin/
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/web ./cmd/web/

FROM alpine:3.19

RUN apk add --no-cache sqlite-libs ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/ ./bin/
COPY --from=builder /app/templates/ ./templates/
COPY --from=builder /app/static/ ./static/

RUN mkdir -p /app/data /app/uploads/products /app/uploads/banners

EXPOSE 8600 18600
