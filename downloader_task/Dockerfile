FROM alpine:3.11
RUN apk add --no-cache wget
RUN apk add --no-cache  aws-cli --repository http://dl-3.alpinelinux.org/alpine/edge/community/

WORKDIR /tmp/workdir

COPY upload.sh /tmp/workdir

ENTRYPOINT wget https://atp.fm:8443/listen -O `date --utc +%s`.mp3 && ./upload.sh
