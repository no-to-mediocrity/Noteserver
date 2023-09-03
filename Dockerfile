# syntax=docker/dockerfile:1

FROM golang:1.19
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /go/bin/noteserver ./cmd/noteserver/main.go
WORKDIR /
CMD go/bin/noteserver
EXPOSE 5432  
EXPOSE 8080