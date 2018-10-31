FROM alpine:3.7
RUN apk add --no-cache wget
RUN apk add aws-cli --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/

WORKDIR /tmp/workdir

COPY upload.sh /tmp/workdir

ENTRYPOINT wget http://file-examples.com/wp-content/uploads/2017/11/file_example_MP3_1MG.mp3 -O `date --utc +%s`.mp3 && ./upload.sh
