FROM alpine:3.7
RUN apk add --no-cache wget
RUN apk add aws-cli --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/

WORKDIR /tmp/workdir

COPY upload.sh /tmp/workdir

ENTRYPOINT wget https://atp.fm:8443/listen -O `date --utc +%s`.mp3 && ./upload.sh
