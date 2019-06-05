#!/bin/bash

cd /opt/milbot

# milbot-redis のビルド
echo "building milbot-redis"
docker build -t milbot-redis -f redis_docker/Dockerfile_rpi redis_docker
docker run -d --name milbot-redis --memory=100m --memory-swappiness=0 -p 6379:6379 milbot-redis
docker stop milbot-redis
echo "done"

# milbot のビルド
echo "building milbot"
docker build -t milbot -f milbot/Dockerfile_rpi milbot
echo "done"

# Systemd のサービスファイルのインストール
echo "linking systemd service files"
cp systemd/milbot.target /etc/systemd/system/milbot.target
cp systemd/milbot-bluetooth.service /etc/systemd/system/milbot-bluetooth.service
cp systemd/milbot-redis.service /etc/systemd/system/milbot-redis.service
cp systemd/milbot.service /etc/systemd/system/milbot.service
echo "done"