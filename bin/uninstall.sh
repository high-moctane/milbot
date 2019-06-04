#!/bin/bash

cd /opt/milbot

# Systemd のサービスファイルの削除
echo "removing systemd service files"
rm /etc/systemd/bluetooth_server.service
rm /etc/systemd/milbot-redis.service
rm /etc/systemd/bluetooth_server.service
echo "done"

# milbot のコンテナを消す
echo "removing milbot container"
docker rm milbot
echo "done"

# milbot-redis のコンテナを消す
echo "removing milbot-redis container"
docker rm milbot-redis
echo "done"

# milbot のイメージを消す
echo "removing milbot image"
docker rmi milbot
echo "done"

# milbot-redis のイメージを消す
echo "removing milbot-redis image"
docker rmi milbot-redis
echo "done"