FROM alpine:3.7
RUN apk add --no-cache wget
RUN apk add aws-cli --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/

WORKDIR /tmp/workdir

COPY upload.sh /tmp/workdir

ENTRYPOINT wget http://marco.org:8001/listen -O `date --utc +%s`.mp3 && ./upload.sh
