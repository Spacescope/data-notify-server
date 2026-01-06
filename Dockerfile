FROM golang:1.24.7-trixie AS builder

COPY . /opt
RUN cd /opt && go build -o bin/data-extraction-notify cmd/data-extraction-notify/main.go

FROM debian:trixie
RUN apt update && apt-get install ca-certificates -y
RUN useradd -m -d /app/data-notify-server -c "Devops Starboard,Github,WorkPhone,HomePhone" -s /usr/sbin/nologin spacescope
USER spacescope
COPY --from=builder /opt/bin/data-extraction-notify /app/data-notify-server/data-notify-server

CMD ["--conf", "/app/data-notify-server/service.conf"]
ENTRYPOINT ["/app/data-notify-server/data-notify-server"]
