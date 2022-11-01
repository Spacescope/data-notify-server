FROM golang:1.18.3-bullseye as builder

COPY . /opt
RUN cd /opt && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/data-api-server cmd/data-extraction-notify/main.go

FROM alpine:3.15.4
RUN mkdir -p /app/data-extraction-notify-backend
RUN adduser -h /app/data-extraction-notify-backend -D starboard
USER starboard
COPY --from=builder /opt/bin/data-api-server /app/data-extraction-notify-backend/data-api-server

CMD ["--conf", "/app/data-extraction-notify-backend/service.conf"]
ENTRYPOINT ["/app/data-extraction-notify-backend/data-api-server"]
