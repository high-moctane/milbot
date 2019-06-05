#!/bin/bash

cd /opt/milbot

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

echo "removing unit files"
rm /etc/systemd/system/milbot.target
rm /etc/systemd/system/milbot-bluetooth.service
rm /etc/systemd/system/milbot-redis.service
rm /etc/systemd/system/milbot.service
echo "done"
echo ""

echo "pulling milbot"
git fetch origin master
git reset --hard origin/master
echo "done"
echo ""

echo "copying unit files"
cp systemd/milbot.target /etc/systemd/system/milbot.target
cp systemd/milbot-bluetooth.service /etc/systemd/system/milbot-bluetooth.service
cp systemd/milbot-redis.service /etc/systemd/system/milbot-redis.service
cp systemd/milbot.service /etc/systemd/system/milbot.service
echo "done"
echo ""