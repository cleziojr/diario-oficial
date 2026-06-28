FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY backend/go.mod backend/go.sum ./backend/
COPY ai-service/go.mod ai-service/go.sum ./ai-service/
COPY pdf-extraction/go.mod pdf-extraction/go.sum ./pdf-extraction/
RUN cd pdf-extraction && go mod download
COPY pdf-extraction/ ./pdf-extraction/
RUN CGO_ENABLED=0 GOOS=linux go build -o pdf-extraction/pdf-extraction ./pdf-extraction/cmd/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata poppler-utils
WORKDIR /app
COPY --from=builder /app/pdf-extraction/pdf-extraction .
ENTRYPOINT ["./pdf-extraction"]