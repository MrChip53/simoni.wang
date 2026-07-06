FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:3.18
WORKDIR /app
COPY --from=go-builder /app/main .
EXPOSE 8080
CMD ["./main"]