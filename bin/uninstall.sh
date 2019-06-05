#!/bin/bash

cd /opt/milbot

# サービス停止
echo "stopping milbot"
systemctl stop milbot.service
systemctl stop milbot-redis.service
systemctl stop milbot-bluetooth.service
systemctl stop milbot.target
echo "done"
echo ""

echo "disabling milbot"
systemctl disable milbot.target
echo "done"
echo ""

# Systemd のサービスファイルの削除
echo "removing systemd service files"
rm /etc/systemd/system/milbot.target
rm /etc/systemd/system/milbot-bluetooth.service
rm /etc/systemd/system/milbot-redis.service
rm /etc/systemd/system/milbot.service
echo "done"
echo ""

# milbot のコンテナを消す
echo "removing milbot container"
docker rm milbot
echo "done"
echo ""

# milbot-redis のコンテナを消す
echo "removing milbot-redis container"
docker rm milbot-redis
echo "done"
echo ""

# milbot のイメージを消す
echo "removing milbot image"
docker rmi milbot
echo "done"
echo ""

# milbot-redis のイメージを消す
echo "removing milbot-redis image"
docker rmi milbot-redis
echo "done"
echo ""