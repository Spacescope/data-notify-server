FROM golang:1.18.3-bullseye as builder

COPY . /opt
RUN cd /opt && go build -o bin/data-extraction-notify cmd/data-extraction-notify/main.go

FROM debian:bullseye
RUN apt update && apt-get install ca-certificates -y
RUN adduser --gecos "Devops Starboard,Github,WorkPhone,HomePhone" --home /app/data-notify-server --disabled-password spacescope
USER spacescope
COPY --from=builder /opt/bin/data-extraction-notify /app/data-notify-server/data-notify-server

CMD ["--conf", "/app/data-notify-server/service.conf"]
ENTRYPOINT ["/app/data-notify-server/data-notify-server"]
