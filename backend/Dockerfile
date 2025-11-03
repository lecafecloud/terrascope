FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o terrascope ./cmd/api/

FROM alpine:3.22 AS final
RUN apk --no-cache add ca-certificates=20250911-r0
WORKDIR /app
COPY --from=builder /app/terrascope ./
EXPOSE 8080
ENTRYPOINT [ "./terrascope" ]
