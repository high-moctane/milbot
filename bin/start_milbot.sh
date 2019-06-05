#!/bin/bash

ip_addr=$(ip addr show eth0 | grep -o 'inet [0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+' | grep -o [0-9].*)
/usr/bin/docker run --pids-limit 30 --ulimit fsize=100000000:100000000 --init --name milbot --rm --memory=250m --memory-swappiness=0 --add-host=host_address:${ip_addr} milbot