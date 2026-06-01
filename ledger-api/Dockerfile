FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod ./
RUN go env -w GONOSUMDB="*" && go mod download

COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ledger-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ledger-api /ledger-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ledger-api"]
