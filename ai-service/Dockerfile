FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY ai-service/go.mod ai-service/go.sum ./ai-service/
RUN cd ai-service && go mod download
COPY ai-service/ ./ai-service/
RUN cd ai-service && CGO_ENABLED=0 GOOS=linux go build -o ai-service ./cmd/server/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates wget tzdata
WORKDIR /app
COPY --from=builder /app/ai-service/ai-service .
EXPOSE 9090
ENTRYPOINT ["./ai-service"]