FROM balenalib/rpi-debian-python:3.7.3-stretch-build

RUN apt-get update && \
    aptget install -y redis-tools ca-certificates && \
    useradd -m admin
WORKDIR /home/admin
USER admin
COPY milbot-rpi .env ./

CMD ./milbot-rpi
