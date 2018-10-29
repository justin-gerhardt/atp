FROM jrottenberg/ffmpeg

RUN apt-get update && \
    apt-get install python-dev python-pip -y && \
    apt-get clean

RUN pip install awscli

WORKDIR /tmp/workdir

COPY ./copy_thumbs.sh ./

ENTRYPOINT ffmpeg -i http://file-examples.com/wp-content/uploads/2017/11/file_example_MP3_1MG.mp3 && ./copy_thumbs.sh
