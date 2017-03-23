FROM alpine:3.1
MAINTAINER Lewis Watson <mrlewiswatson@gmail.com>
ADD carshare-back /usr/bin/carshare-back
ENTRYPOINT ["carshare-back"]
