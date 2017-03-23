FROM alpine:3.5
MAINTAINER Lewis Watson <mrlewiswatson@gmail.com>
# need to ensure CA Certs are installed so we can securely get firebase keys
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*
ADD carshare-back /usr/bin/carshare-back
ENTRYPOINT ["carshare-back"]
