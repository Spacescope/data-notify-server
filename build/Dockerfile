FROM alpine:3.15

RUN mkdir -p /etc/data-extraction-notify/{bin,conf}
COPY bin/data-api-server /etc/data-extraction-notify/bin/data-api-server

CMD ["--conf", "/etc/data-extraction-notify/conf/service.conf"]
ENTRYPOINT ["/etc/data-extraction-notify/bin/data-api-server"]