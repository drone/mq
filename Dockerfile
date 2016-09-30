FROM centurylink/ca-certs
MAINTAINER Brad Rydzewski <brad.rydzewski@gmail.com>
EXPOSE 8000 9000

ENV GODEBUG=netdns=go
ADD release/linux_amd64/mq /mq

ENTRYPOINT ["/mq"]
CMD ["start"]
