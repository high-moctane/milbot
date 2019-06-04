#!/bin/bash

cd /opt/milbot

# milbot-redis のビルド
echo "building milbot-redis"
docker build -t milbot-redis -f redis_docker/Dockerfile_rpi redis_docker
docker run --name milbot-redis --memory=100m --memory-swappiness=0 -p 6379:6379 milbot-redis
docker stop milbot-redis
echo "done"

# milbot のビルド
echo "building milbot-redis"
docker build -t milbot -f milbot/Dockerfile_rpi milbot
echo "done"

# Systemd のサービスファイルのインストール
echo "linking systemd service files"
ln -s systemd/bluetooth_server.service /etc/systemd/bluetooth_server.service
ln -s systemd/milbot-redis.service /etc/systemd/milbot-redis.service
ln -s systemd/bluetooth_server.service /etc/systemd/bluetooth_server.service
echo "done"

# service の有効化
echo "enabling systemd services"
systemctl enable bluetooth_server
systemctl enable milbot-redis
systemctl enable milbot
echo "done"