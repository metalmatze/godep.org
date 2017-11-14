FROM alpine
EXPOSE 8000

RUN apk add --update ca-certificates
ADD ./godep.org /usr/bin/godep.org

ENTRYPOINT ["/usr/bin/godep.org"]
