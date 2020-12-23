#!/bin/bash
set -e

ulimit -n 1000000

DEPLOY_DIR=/home/david/Documents/project/cubit
cd "${DEPLOY_DIR}" || exit 1

exec docker run \
        -p 3306:3306 \
        -v /etc/localtime:/etc/localtime:ro \
        -v mysql_database:/var/lib/mysql \
        -v "${DEPLOY_DIR}/data/mysql-entrypoint:/docker-entrypoint-initdb.d" \
        -e MYSQL_USER=mysql \
        -e MYSQL_PASSWORD=password \
        -e MYSQL_ROOT_PASSWORD=secret \
        -e TZ=Asia/Shanghai \
        --restart on-failure \
        --name cubit_mysql \
        mysql:8.0.21
