#!/bin/bash

cd /opt/milbot

# Systemd のサービスファイルの削除
echo "removing systemd service files"
rm /etc/systemd/bluetooth_server.service
rm /etc/systemd/milbot-redis.service
rm /etc/systemd/bluetooth_server.service
echo "done"

# milbot のコンテナを消す
if [ $(docker ps -a | grep milbot) ]
then
    echo "removing milbot container"
    docker rm milbot
    echo "done"
fi
# milbot-redis のコンテナを消す
if [ $(docker ps -a | grep milbot-redis) ]
then
    echo "removing milbot-redis container"
    docker rm milbot-redis
    echo "done"
fi

# milbot のイメージを消す
if [ $(docker images | grep milbot) ]
then
    echo "removing milbot image"
    docker rmi milbot
    echo "done"
fi
# milbot-redis のイメージを消す
if [ $(docker images | grep milbot-redis) ]
then
    echo "removing milbot-redis image"
    docker rmi milbot-redis
    echo "done"
fi