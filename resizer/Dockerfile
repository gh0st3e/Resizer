FROM golang:alpine AS builder

COPY resizer/. /src
WORKDIR /src
RUN go build -o main.bin cmd/main.go

FROM alpine as runner

WORKDIR /app
COPY --from=builder /src/main.bin /app/

ENTRYPOINT ./main.bin

CMD ["/app"]